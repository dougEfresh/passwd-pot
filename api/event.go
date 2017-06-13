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

package api

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const EventURL = "/api/v1/event"
const EventCountryStatsUrl = "/api/v1/event/stats/country"
const StreamURL = "/api/v1/event/stream"

//Custom Serializer
type EventTime time.Time

// Time is in epoch ms
func (et *EventTime) UnmarshalJSON(data []byte) (err error) {
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return errors.New(fmt.Sprintf("could not decode time %s err:%s", data, err))
	}
	*et = EventTime(time.Unix(ts/1000, (ts%1000)*1000000).UTC())
	return nil
}

func (et EventTime) MarshalJSON() ([]byte, error) {
	ts := time.Time(et).UTC().UnixNano() / int64(time.Millisecond)
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

// Value implements the driver Valuer interface.
func (et EventTime) Value() (driver.Value, error) {
	return time.Time(et), nil
}

// Gets the value from epoch time
func (et *EventTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		return nil
	case []byte:
		et.UnmarshalJSON(v)
		return nil
	case string:
		et.UnmarshalJSON([]byte(v))
		return nil
	}
	return nil
}

type EventClient struct {
	server string
}

func (e *EventClient) RecordEvent(event Event) (int64, error) {
	b, err := json.Marshal(event)
	if err != nil {
		return 0, err
	}
	var body []byte
	err = backoff.Retry(func() error {
		body, err = e.transport("POST", EventURL, b)
		return err
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return 0, err
	}
	return 0, nil
}

//TODO
func (e *EventClient) GetEvent(id int64) (*EventGeo, error) {
	return nil, nil
}

func (e *EventClient) GetCountryStats() ([]CountryStat, error) {
	var stats []CountryStat
	resp, err := e.transport("GET", EventCountryStatsUrl, nil)
	if err != nil {
		return stats, err
	}
	err = json.Unmarshal(resp, stats)
	return stats, err
}

func (e *EventClient) transport(method string, endpoint string, body []byte) ([]byte, error) {
	var res *http.Response
	var err error
	if method == "POST" {
		res, err = http.Post(fmt.Sprintf("%s%s", e.server, endpoint),
			"application/json",
			bytes.NewReader(body))
	} else {
		res, err = http.Get(fmt.Sprintf("%s%s", e.server, endpoint))
	}

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)

	if res.StatusCode == http.StatusAccepted || res.StatusCode == http.StatusOK {
		return b, nil
	}
	return nil, errors.New(fmt.Sprintf("Something when wrong http status code: %d", res.StatusCode))

}

func NewClient(server string, options ...func(*EventClient) error) (*EventClient, error) {
	ec := &EventClient{
		server: server,
	}

	for _, opt := range options {
		if err := opt(ec); err != nil {
			return nil, err
		}
	}
	return ec, nil
}

func (et EventTime) String() string {
	return fmt.Sprintf("%s", time.Time(et))
}

func (e Event) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return string(b)
}
