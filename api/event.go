package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const EventURL = "/api/v1/event"
const StreamURL = "/api/v1/event/stream"

//Custom Serializer
type EventTime time.Time

// Time is in epoch ms
func (et *EventTime) UnmarshalJSON(data []byte) (err error) {
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return errors.New("could not decode time " + string(data))
	}
	*et = EventTime(time.Unix(ts/1000, (ts%1000)*1000000).UTC())
	return nil
}

func (et *EventTime) MarshalJSON() ([]byte, error) {
	ts := time.Time(*et).UTC().UnixNano() / int64(time.Millisecond)
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

//Event to record
type Event struct {
	ID            int64
	Time          EventTime
	User          string
	Passwd        string
	RemoteAddr    string
	RemotePort    int
	RemoteName    string
	RemoteVersion string
	OriginAddr    string
	Application   string
	Protocol      string
}

type Transporter interface {
	sendEvent(event *Event) *Event
	getEvent(id int64) *Event
}

type EventClient struct {
	server string
}

func (e *EventClient) sendEvent(event *Event) (*Event, error) {

	b, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	body, err := e.transport("POST", EventURL, b)

	if err != nil {
		return nil, err
	}
	return convert(body)
}

//TODO
func (e *EventClient) getEvent(id int64) (*EventClient, error) {
	return nil, nil
}

func convert(b []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(b, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (e *EventClient) transport(method string, endpoint string, body []byte) ([]byte, error) {
	var res *http.Response
	var err error
	if method == "POST" {
		res, err = http.Post(fmt.Sprintf("%s%s", e.server, endpoint),
			"application/json",
			bytes.NewReader(body))
	} else {
		res, err = http.Post(fmt.Sprintf("%s%s", e.server, endpoint),
			"application/json",
			bytes.NewReader(body))
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
