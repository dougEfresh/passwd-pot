package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	awsevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/dougEfresh/lambdazap"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/event"
	"github.com/dougEfresh/passwd-pot/potdb"
	"github.com/dougEfresh/passwd-pot/resolver"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultDsn = "postgres://postgres:@127.0.0.1/?sslmode=disable"

var (
	eventResolver *resolver.ResolveClient
	eventClient   *event.Client
	dsn           = os.Getenv("PASSWDPOT_DSN")
	geoServer     = os.Getenv("PASSWDPOT_GEO_SERVER")
	setupError    error
	db            potdb.DB
	geoClient *resolver.GeoClient
)

var header = map[string]string{
	"Content-Type": "application/json",
}

var logger *zap.Logger

func setup() {
	if dsn == "" {
		dsn = defaultDsn
	}
	var err error
	if db == nil {
		db, _ = potdb.Open(dsn)
	}

	if err = db.Ping(); err != nil {
		logger.Error(fmt.Sprintf("Error loading db %s", err))
		setupError = err
		return
	}
	event.SetDB(db)(eventClient)
	resolver.SetDb(db)(eventResolver)
}

func init() {
	logger, _ = zap.NewProduction()
	eventClient, _ = event.New()
	geoClient = &resolver.GeoClient {
		URL:"http://geo.passwd-pot.io:8080",
	}
	if geoServer != "" {
		geoClient.URL = geoServer
	}
	eventResolver, _ = resolver.NewResolveClient(resolver.UseCache(), resolver.SetGeoClient(geoClient))
	setup()
}

// APIError Custom error msg
type APIError struct {
	GatewayError awsevents.APIGatewayProxyResponse
}

func (e APIError) Error() string {
	body, err := json.Marshal(e.GatewayError)
	if err != nil {
		return fmt.Sprintf(`{"statusCode": 500, "body": "fatal error!  %s"}`, err)
	}
	return string(body)
}

func (e APIError) String() string {
	return e.Error()
}

// EventResponse Send the ID
type EventResponse struct {
	ID int64 `json:"id"`
}

// BatchEvent API gw
type BatchEvent struct {
	OriginAddr string      `json:"originAddr"`
	Events     []api.Event `json:"events"`
}

var lc = lambdazap.New().WithBasic()

func logThis(ctx context.Context, level zapcore.Level, msg string, args ...interface{}) {
	switch level {
	case zap.ErrorLevel:
		logger.Error(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	case zap.WarnLevel:
		logger.Warn(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	case zap.DebugLevel:
		logger.Debug(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	default:
		logger.Info(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	}

}

func resolveAddr(ctx context.Context, addr string) (int64, error) {
	rID, err := eventResolver.Resolve(addr)
	if err != nil {
		logThis(ctx, zapcore.ErrorLevel, "error resolving %s (%s)", addr, err)
		return 0, APIError{GatewayError: awsevents.APIGatewayProxyResponse{Body: "Batch insert error", StatusCode: 500}}
	}
	return rID, nil
}

// HandleBatch from GW
func HandleBatch(ctx context.Context, events BatchEvent) (api.BatchEventResponse, error) {
	defer logger.Sync()
	var geoIds = make(map[string]int64)
	// Stupid API GW
	if events.OriginAddr == "test-invoke-source-ip" {
		events.OriginAddr = "127.0.0.1"
	}
	var id int64
	if setupError != nil {
		setup()
		return api.BatchEventResponse{}, APIError{GatewayError: awsevents.APIGatewayProxyResponse{StatusCode: 500, Headers: header, Body: fmt.Sprintf("SetupError %s", setupError)}}
	}
	gid, err := resolveAddr(ctx, events.OriginAddr)
	if err != nil {
		return api.BatchEventResponse{}, err
	}
	for i := 0; i < len(events.Events); i++ {
		e := &events.Events[i]
		if e.RemoteAddr == "" {
			return api.BatchEventResponse{}, APIError{GatewayError: awsevents.APIGatewayProxyResponse{StatusCode: 500, Headers: header, Body: fmt.Sprintf("Error event %s", e)}}
		}
		_, ok := geoIds[e.RemoteAddr]
		if !ok {
			id, err = resolveAddr(ctx, e.RemoteAddr)
			if err != nil {
				return api.BatchEventResponse{}, err
			}
			geoIds[e.RemoteAddr] = id
		}
		e.RemoteGeoID = id
		e.OriginGeoID = gid
		e.OriginAddr = events.OriginAddr
	}

	resp, err := eventClient.RecordBatchEvents(events.Events)
	if err != nil {
		return api.BatchEventResponse{}, APIError{GatewayError: awsevents.APIGatewayProxyResponse{StatusCode: 500, Headers: header, Body: fmt.Sprintf("Error with batch insert %s", err)}}
	}
	logger.Info("result", zap.Int64("duration", resp.Duration.Nanoseconds()), zap.Int64("rows", resp.Rows))
	logThis(ctx, zapcore.InfoLevel, "batchDuration:%d rows:%d", resp.Duration, resp.Rows)
	return resp, nil
}

// Handle Password Event
func Handle(ctx context.Context, e api.Event) (EventResponse, error) {

	defer logger.Sync()
	if e.RemoteAddr == "" {
		e := APIError{awsevents.APIGatewayProxyResponse{Body: fmt.Sprintf("error with event %s", e), StatusCode: 400}}
		logThis(ctx, zapcore.ErrorLevel, "error with event %s", e)
		return EventResponse{}, e
	}
	if setupError != nil {
		resp := APIError{awsevents.APIGatewayProxyResponse{Body: fmt.Sprintf("Bad setup %s", setupError), StatusCode: 500}}
		logThis(ctx, zapcore.ErrorLevel, "%s", resp)
		setupError = nil
		setup()
		return EventResponse{}, resp
	}
	logThis(ctx, zapcore.DebugLevel, "Event %s", e)
	if e.OriginAddr == "test-invoke-source-ip" {
		// Stupid API gw
		e.OriginAddr = "127.0.0.1"
	}
	id, err := eventClient.RecordEvent(e)
	if err != nil {
		logThis(ctx, zapcore.ErrorLevel, "error loading event %s", err)
		return EventResponse{}, APIError{awsevents.APIGatewayProxyResponse{Body: fmt.Sprintf("error loading event %s", err), StatusCode: 500}}
	}
	e.ID = id
	_, err = eventResolver.ResolveEvent(e)
	if err != nil {
		logThis(ctx, zapcore.WarnLevel, "Error resolving %s or %s %s", e.RemoteAddr, e.OriginAddr, err)
	}
	return EventResponse{id}, nil
}

func main() {
	if strings.Contains(lambdacontext.FunctionName, "batch") {
		lambda.Start(HandleBatch)
	} else if strings.Contains(lambdacontext.FunctionName, "geolookup") {
		lambda.Start(HandleGeoLookup)
	} else {
		lambda.Start(Handle)
	}
}
