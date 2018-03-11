package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
	"github.com/dougEfresh/passwd-pot/service"
	klog "github.com/go-kit/kit/log"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

var eventResolver service.EventResolver
var eventClient *service.EventClient
var logger log.Logger
var dsn = os.Getenv("PASSWDPOT_DSN")
var setupError error

func init() {
	logger.SetLevel(log.InfoLevel)
	logger.AddLogger(klog.NewJSONLogger(os.Stdout))
	logger.With("app", "passwdpot-create-event")
	logger.With("ts", klog.DefaultTimestampUTC)
	logger.With("caller", klog.Caller(4))
	var err error
	if os.Getenv("PASSWDPOT_DEBUG") == "1" {
		logger.SetLevel(log.DebugLevel)
	}
	eventResolver, err = service.NewResolveClient(service.WithResolvDsn(dsn))
	if err != nil {
		logger.Errorf("Error setting up client %s", err)
		setupError = errors.New(fmt.Sprintf("resolver has bad setup %s", err))
		return
	}
	eventClient, err = service.NewEventClient(service.SetEventLogger(logger), service.WithDsn(dsn))
	if err != nil {
		logger.Errorf("Error setting up eventClient %s", err)
		setupError = errors.New(fmt.Sprintf("eventClient  has bad setup %s", err))
	}
}

type ApiEvent struct {
	Event api.Event `json:"event"`
}

func Handle(apiEvent ApiEvent) (events.APIGatewayProxyResponse, error) {
	e := apiEvent.Event
	if e.RemoteAddr == "" {
		return events.APIGatewayProxyResponse{Body: fmt.Sprintf("error  with event %s", e), StatusCode: 500}, errors.New("Invalid event")
	}
	if setupError != nil {
		return events.APIGatewayProxyResponse{Body: "Bad setup", StatusCode: 500}, setupError
	}
	logger.Debugf("Event %s", e)
	id, err := eventClient.RecordEvent(e)
	if err != nil {
		logger.Errorf("error loading event %s", err)
		return events.APIGatewayProxyResponse{Body: fmt.Sprintf("error loading event %s", err), StatusCode: 500}, err
	}
	e.ID = id
	_, err = eventResolver.ResolveEvent(e)
	if err != nil {
		logger.Errorf("Error resolving %s %s", e, err)
		return events.APIGatewayProxyResponse{Body: fmt.Sprintf("error resolving event %s", err), StatusCode: 500}, err
	}
	var header = make(map[string]string)
	header["Content-Type"] = "application/json;charset=UTF-8"
	return events.APIGatewayProxyResponse{Body: fmt.Sprintf("{\"id\":%d}", id), StatusCode: 202, Headers: header}, nil
}

func main() {
	lambda.Start(Handle)
}
