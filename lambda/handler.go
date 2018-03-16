package main

import (
	"database/sql"
	"errors"
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
)

const DEFAULT_DSN = "root@tcp(127.0.0.1:3306)/passwdpot?tls=false&parseTime=true&loc=UTC&timeout=10ms"

var geoCache *cache.Cache = cache.NewCache()
var eventResolver service.EventResolver
var eventClient *service.EventClient
var logger log.Logger
var dsn = os.Getenv("PASSWDPOT_DSN")
var logz = os.Getenv("LOGZ")
var setupError error
var db *sql.DB

func loadDSN(dsn string) error {
	var err error
	if strings.Contains(dsn, "postgres") {
		logger.Debug("Using pq driver")
		db, err = sql.Open("postgres", dsn)
	} else {
		logger.Debug("Using mysql driver")
		db, err = sql.Open("mysql", dsn)
	}

	if err != nil {
		return err
	}
	return db.Ping()
}

var header = map[string]string{
	"Content-Type": "application/json",
}

func init() {
	if dsn == "" {
		dsn = DEFAULT_DSN
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
	err = loadDSN(dsn)
	if err != nil {
		logger.Errorf("Error loading db %s", err)
		setupError = err
		return
	}
	eventResolver, err = service.NewResolveClient(service.SetResolveDb(db), service.SetResolveLogger(logger))
	if err != nil {
		logger.Errorf("Error setting up client %s", err)
		setupError = errors.New(fmt.Sprintf("resolver has bad setup %s", err))
		return
	}
	eventClient, err = service.NewEventClient(service.SetEventLogger(logger), service.SetEventDb(db))
	if err != nil {
		logger.Errorf("Error setting up eventClient %s", err)
		setupError = errors.New(fmt.Sprintf("eventClient  has bad setup %s", err))
	}
}

// ApiEvent from API GW
type ApiEvent struct {
	Event api.Event `json:"event"`
}
type EventError struct {
	 Event api.Event
	 Msg string
}

func (e EventError) Error() string {
	return fmt.Sprintf("error with event: %s\nmsg: %s" ,e.Event, e.Msg )
}
func (e EventError) String() string {
	return e.Error()
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

func checkDB() bool {
	if db == nil {
		if err := loadDSN(dsn) ; err != nil {
			return false
		}
	}
	if err := db.Ping(); err != nil {
		return false
	}
	return true
}

func sendError(e EventError) (events.APIGatewayProxyResponse, error) {
	logger.Errorf("%s", e)
	return events.APIGatewayProxyResponse{Body: fmt.Sprintf("%s", e), StatusCode: 500}, e
}

// Handle Password Event
func Handle(apiEvent ApiEvent) (events.APIGatewayProxyResponse, error) {
	e := apiEvent.Event
	if !checkDB() {
		return sendError(EventError{Event:e, Msg:"db is not set up"})
	}
	if e.RemoteAddr == "" {
		return sendError(EventError{Event:e, Msg:"invalid event"})
	}
	logger.Debugf("Event %s", e)
	id, err := eventClient.RecordEvent(e)
	if err != nil {
		return sendError(EventError{Event:e, Msg: fmt.Sprintf("could not record event %s", err)})
	}
	e.ID = id
	err = resolveEvent(e)
	if err != nil {
		logger.Warnf("Error resolving %s %s", e, err)
	}
	return events.APIGatewayProxyResponse{Body: fmt.Sprintf("{id:%d}", id), StatusCode: 202, Headers: header}, nil
}

func main() {
	lambda.Start(Handle)
}
