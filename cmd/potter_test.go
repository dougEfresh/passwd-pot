// Copyright © 2017 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
)

type mockEventClient struct {
}

var submittedEvent *api.Event

func (c *mockEventClient) RecordEvent(event api.Event) (int64, error) {
	submittedEvent = &event
	return 0, nil
}

func (c *mockEventClient) RecordBatchEvents(event []api.Event) (api.BatchEventResponse, error) {
	return api.BatchEventResponse{}, nil
}

func (c *mockEventClient) GetEvent(id int64) (*api.EventGeo, error) {
	return &api.EventGeo{}, nil
}

func (c *mockEventClient) GetCountryStats() ([]api.CountryStat, error) {
	return nil, nil
}

func init() {
	logger = log.DefaultLogger(os.Stdout)
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
