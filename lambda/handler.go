package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dougEfresh/lambdazap"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/event"
	"github.com/dougEfresh/passwd-pot/potdb"
	"github.com/dougEfresh/passwd-pot/resolver"
	"github.com/dougEfresh/zapz"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultDsn = "postgres://postgres:@127.0.0.1/?sslmode=disable"

var (
	eventResolver *resolver.ResolveClient
	eventClient   *event.EventClient
	dsn           = os.Getenv("PASSWDPOT_DSN")
	logz          = os.Getenv("LOGZ")
	setupError    error
	db            potdb.DB
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
	event.SetEventDb(db)(eventClient)
	resolver.SetDb(db)(eventResolver)
}

func init() {
	var err error
	logger, _ = zap.NewProduction()
	if logz != "" {
		logger, err = zapz.New(logz)
		if err != nil {
			fmt.Sprintf("Error loading logz %s", err)
		}
	}
	eventClient, _ = event.NewEventClient()
	eventResolver, _ = resolver.NewResolveClient(resolver.UseCache())
	setup()
}

// APIError Custom error msg
type APIError struct {
	GatewayError events.APIGatewayProxyResponse
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

func sendError(e events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	body, err := json.Marshal(e)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("%s", err),
			Headers:    header,
		}
	}
	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: e.StatusCode,
		Headers:    header,
	}
}

// EventResponse Send the ID
type EventResponse struct {
	ID int64 `json:"id"`
}

var lc = lambdazap.New().WithBasic()

func logThis(ctx context.Context, level zapcore.Level, msg string, args ...interface{}) {
	switch level {
	case zap.ErrorLevel:
		logger.Error(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	case zap.WarnLevel:
		logger.Error(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	case zap.DebugLevel:
		logger.Debug(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	default:
		logger.Info(fmt.Sprintf(msg, args...), lc.ContextValues(ctx)...)
	}

}

// Handle Password Event
func Handle(ctx context.Context, e api.Event) (EventResponse, error) {

	defer logger.Sync()
	if e.RemoteAddr == "" {
		e := APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("error with event %s", e), StatusCode: 400}}
		logThis(ctx, zapcore.ErrorLevel, "error with event %s", e)
		return EventResponse{}, e
	}
	if setupError != nil {
		resp := APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("Bad setup %s", setupError), StatusCode: 500}}
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
		return EventResponse{}, APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("error loading event %s", err), StatusCode: 500}}
	}
	e.ID = id
	_, err = eventResolver.ResolveEvent(e)
	if err != nil {
		logThis(ctx, zapcore.ErrorLevel, "Error resolving %s %s", e, err)
	}
	return EventResponse{id}, nil
}

func main() {
	lambda.Start(Handle)
}
