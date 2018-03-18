package main

import (
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
	klog "github.com/go-kit/kit/log"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const defaultDsn = "postgres://postgres:@127.0.0.1/?sslmode=disable"

var eventResolver resolver.EventResolver
var eventClient *event.EventClient
var logger log.Logger
var dsn = os.Getenv("PASSWDPOT_DSN")
var logz = os.Getenv("LOGZ")
var setupError error

var header = map[string]string{
	"Content-Type": "application/json",
}

func init() {
	if dsn == "" {
		dsn = defaultDsn
	}

	logger.AddLogger(klog.NewJSONLogger(os.Stdout))
	logger.With("app", "passwdpot-create-event")
	logger.With("ts", klog.DefaultTimestampUTC)
	logger.With("caller", klog.Caller(4))
	if logz != "" {
		lz, err := kitz.New(logz)
		if err != nil {
			logger.Errorf("Error connecting to logz %s\n", err)
		} else {
			logger.AddLogger(lz)
		}
	}
	if os.Getenv("PASSWDPOT_DEBUG") == "1" {
		logger.SetLevel(log.DebugLevel)
	}
	var err error
	db, err := potdb.Open(dsn)

	if err != nil {
		logger.Errorf("Error loading db %s", err)
		setupError = err
		return
	}
	eventResolver, err = resolver.NewResolveClient(resolver.SetDb(db), resolver.SetLogger(logger), resolver.UseCache())
	if err != nil {
		logger.Errorf("Error setting up client %s", err)
		setupError = fmt.Errorf("resolver has bad setup %s", err)
		return
	}
	eventClient, err = event.NewEventClient(event.SetEventLogger(logger), event.SetEventDb(db))
	if err != nil {
		logger.Errorf("Error setting up eventClient %s", err)
		setupError = fmt.Errorf("eventClient  has bad setup %s", err)
	}
}

// APIError Custom error msg
type APIError struct {
	GatewayError events.APIGatewayProxyResponse
}

func (e APIError) Error() string {
	body, err := json.Marshal(e.GatewayError)
	if err != nil {
		return `{"statusCode": 500, "body": "fatal error!"}`
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

// Handle Password Event
func Handle(e api.Event) (EventResponse, error) {
	if e.RemoteAddr == "" {
		return EventResponse{}, APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("error with event %s", e), StatusCode: 400}}
	}
	if setupError != nil {
		return EventResponse{}, APIError{events.APIGatewayProxyResponse{Body: "Bad setup", StatusCode: 500}}
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
