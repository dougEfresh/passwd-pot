package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dougEfresh/kitz"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/event"
	"github.com/dougEfresh/passwd-pot/log"
	"github.com/dougEfresh/passwd-pot/potdb"
	"github.com/dougEfresh/passwd-pot/resolver"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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

func newLogger(ctx context.Context) log.FieldLogger {
	var logger = &log.Logger{}
	/*
		kl := lambdalogcontext.Build(klog.NewJSONLogger(os.Stdout), ctx).WithBasic().Logger()
		logger.AddLogger(kl)
		logger.With("ts", klog.DefaultTimestampUTC)
		logger.With("caller", klog.Caller(4))
		if logz != "" {
			lz, err := kitz.New(logz)
			if err != nil {
				logger.Errorf("Error connecting to logz %s\n", err)
			} else {
				logger.AddLogger(lambdalogcontext.Build(lz, ctx).WithBasic().Logger())
			}
		}
		if os.Getenv("PASSWDPOT_DEBUG") == "1" {
			logger.SetLevel(log.DebugLevel)
		}
	*/
	return logger
}

var defaultLogger = log.DefaultLogger(os.Stdout)

func setup() {
	if dsn == "" {
		dsn = defaultDsn
	}
	var err error
	if db == nil {
		db, _ = potdb.Open(dsn)
	}

	if err = db.Ping(); err != nil {
		defaultLogger.Errorf("Error loading db %s", err)
		setupError = err
		return
	}
	event.SetEventDb(db)(eventClient)
	resolver.SetDb(db)(eventResolver)
}

func init() {
	if logz != "" {
		lz, err := kitz.New(logz)
		if err != nil {
			defaultLogger.Errorf("Error connecting to logz %s\n", err)
		} else {
			defaultLogger.AddLogger(lz)
		}
	}
	eventClient, _ = event.NewEventClient(event.SetEventLogger(defaultLogger))
	eventResolver, _ = resolver.NewResolveClient(resolver.SetLogger(defaultLogger), resolver.UseCache())
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

func getLogger(ctx context.Context) log.FieldLogger {
	return newLogger(ctx)
}

// Handle Password Event
func Handle(ctx context.Context, e api.Event) (EventResponse, error) {
	logger := defaultLogger
	defer logger.Drain()
	//resolver.SetLogger(logger)(eventResolver)
	//event.SetEventLogger(logger)(eventClient)

	if e.RemoteAddr == "" {
		return EventResponse{}, APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("error with event %s", e), StatusCode: 400}}
	}
	if setupError != nil {
		logger.Errorf("Setup is bad %s", setupError)
		resp := APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("Bad setup %s", setupError), StatusCode: 500}}
		setupError = nil
		setup()
		return EventResponse{}, resp
	}
	logger.Debugf("Event %s", e)
	if e.OriginAddr == "test-invoke-source-ip" {
		// Stupid API gw
		e.OriginAddr = "127.0.0.1"
	}
	id, err := eventClient.RecordEvent(e)
	if err != nil {
		logger.Errorf("error loading event %s", err)
		return EventResponse{}, APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("error loading event %s", err), StatusCode: 500}}
	}
	e.ID = id
	_, err = eventResolver.ResolveEvent(e)
	if err != nil {
		logger.Warnf("Error resolving %s %s", e, err)
	}
	return EventResponse{id}, nil
}

func main() {
	lambda.Start(Handle)
}
