package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dougEfresh/kitz"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cache"
	"github.com/dougEfresh/passwd-pot/log"
	"github.com/dougEfresh/passwd-pot/service"
	klog "github.com/go-kit/kit/log"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const defaultDsn = "postgres://postgres:@127.0.0.1/?sslmode=disable"

var geoCache = cache.NewCache()
var eventResolver service.EventResolver
var eventClient *service.EventClient
var logger log.Logger
var dsn = os.Getenv("PASSWDPOT_DSN")
var logz = os.Getenv("LOGZ")
var setupError error

func loadDSN(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	if strings.Contains(dsn, "postgres") {
		logger.Debug("Using pq driver")
		db, err = sql.Open("postgres", dsn)
	} else {
		logger.Debug("Using mysql driver")
		db, err = sql.Open("mysql", dsn)
	}

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, nil
}

var header = map[string]string{
	"Content-Type": "application/json",
}

func init() {
	if dsn == "" {
		dsn = defaultDsn
	}

	logger.SetLevel(log.InfoLevel)
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
	db, err := loadDSN(dsn)

	if err != nil {
		logger.Errorf("Error loading db %s", err)
		setupError = err
		return
	}
	eventResolver, err = service.NewResolveClient(service.SetResolveDb(db), service.SetResolveLogger(logger))
	if err != nil {
		logger.Errorf("Error setting up client %s", err)
		setupError = fmt.Errorf("resolver has bad setup %s", err)
		return
	}
	eventClient, err = service.NewEventClient(service.SetEventLogger(logger), service.SetEventDb(db))
	if err != nil {
		logger.Errorf("Error setting up eventClient %s", err)
		setupError = fmt.Errorf("eventClient  has bad setup %s", err)
	}
}

func resolveEvent(event api.Event) error {
	var ids []int64
	rId, _ := geoCache.Get(event.RemoteAddr)
	oId, _ := geoCache.Get(event.OriginAddr)
	if rId > 0 && oId > 0 {
		if e := eventResolver.MarkRemoteEvent(event.ID, rId); e != nil {
			return e
		}
		if e := eventResolver.MarkOriginEvent(event.ID, oId); e != nil {
			return e
		}
	}
	ids, err := eventResolver.ResolveEvent(event)
	if err == nil {
		geoCache.Set(event.RemoteAddr, ids[0])
		geoCache.Set(event.OriginAddr, ids[1])
	}
	return err
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
	id, err := eventClient.RecordEvent(e)
	if err != nil {
		logger.Errorf("error loading event %s", err)
		return EventResponse{}, APIError{events.APIGatewayProxyResponse{Body: fmt.Sprintf("error loading event %s", err), StatusCode: 500}}
	}
	e.ID = id
	err = resolveEvent(e)
	if err != nil {
		logger.Warnf("Error resolving %s %s", e, err)
	}
	return EventResponse{id}, nil
}

func main() {
	lambda.Start(Handle)
}
