package main

import (
	"encoding/json"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"log"
	"os"
)

type lambdaClient struct {
	eventClient api.Transporter
}

func (lc *lambdaClient) Send(event *api.Event) {
	log.Printf("Sending %s\n", event)
	//lc.eventClient.SendEvent(event)
}

func (lc *lambdaClient) GetEvent(id int64) (*api.Event, error) {
	return nil, nil
}

func Handle(evt json.RawMessage, ctx *runtime.Context) (string, error) {
	var e api.Event
	err := json.Unmarshal(evt, &e)
	server := os.Getenv("API_SERVER")
	log.Printf("Using %s as server", server)
	c, err := api.NewClient(server)
	if err != nil {
		return "ERROR", err
	}
	lc := lambdaClient{c}
	lc.Send(&e)
	return "OK", err
}

func main() {

}
