package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/dougEfresh/passwd-pot/api"
	"testing"
	"time"
)

type mockEventClient struct {
}

var submittedEvent *api.Event

func (c *mockEventClient) SendEvent(event *api.Event) (*api.Event, error) {
	submittedEvent = event
	return event, nil
}

func (c *mockEventClient) GetEvent(id int64) (*api.Event, error) {
	return &api.Event{}, nil
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestPotterClient_Send(t *testing.T) {
	pc := &potterClient{
		eventClient: &mockEventClient{},
	}

	e := &api.Event{
		ID:     1,
		User:   "blah",
		Passwd: "changeme",
	}

	pc.Send(e)
	time.Sleep(250 * time.Millisecond)
	if submittedEvent == nil {
		t.Fatalf("Event never submiited %s", e)
	}

	if submittedEvent.User != "blah" {
		t.Fatalf("Proper event not submitted %s\n%s", e, submittedEvent)
	}

	if submittedEvent.Passwd != "changeme" {
		t.Fatalf("Proper event not submitted %s\n%s", e, submittedEvent)
	}
}
