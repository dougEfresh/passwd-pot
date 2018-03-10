package main

import (
	"encoding/json"
	"fmt"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/service"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
)

func Handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var e api.Event
	if err := json.Unmarshal([]byte(request.Body), &e); err != nil {
		return events.APIGatewayProxyResponse{Body: fmt.Sprintf("%s", err), StatusCode: 403}, nil
	}
	e.ID = 1
	var header = make(map[string]string)
	header["Content-Type"] = "application/json;charset=UTF-8"
	return events.APIGatewayProxyResponse{Body: fmt.Sprintf("{\"id\":%d}", e.ID), StatusCode: 202, Headers: header}, nil

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
	lambda.Start(Handle)
}
