package main

import (
	"encoding/json"
	"fmt"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/service"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

type lambdaClient struct {
}

func Handle(evt json.RawMessage, ctx *runtime.Context) (string, error) {
	var e api.Event
	err := json.Unmarshal(evt, &e)
	db := os.Getenv("DSN")
	c, err := service.NewEventClient(service.WithDsn(db))
	if err != nil {
		log.Printf("Error writing %s", err)
		return "ERROR", err
	}
	id, err := c.RecordEvent(e)
	return fmt.Sprintf("%d", id), err
}

func ResolveEvent(evt json.RawMessage, ctx *runtime.Context) (string, error) {
	var e api.Event
	err := json.Unmarshal(evt, &e)
	db := os.Getenv("DSN")
	c, err := service.NewResolveClient(service.WithResolvDsn(db))
	_, err = c.ResolveEvent(e)
	if err != nil {
		return "ERROR", err
	}
	return "OK", err
}

func main() {

}
