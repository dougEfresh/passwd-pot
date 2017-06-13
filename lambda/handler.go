package main

import (
	"encoding/json"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"log"
	"os"
)

type lambdaClient struct {
}

func Handle(evt json.RawMessage, ctx *runtime.Context) (string, error) {
	var e api.Event
	err := json.Unmarshal(evt, &e)
	server := os.Getenv("API_SERVER")
	log.Printf("Using %s as server", server)
	_, err = api.NewClient(server)
	if err != nil {
		return "ERROR", err
	}
	return "OK", err
}

func main() {

}
