package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/dougEfresh/passwd-pot/api"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	requestBody       = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "1.2.3.4", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "application": "OpenSSH" , "protocol": "ssh"}`
	requestBodyOrigin = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "192.168.1.1", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "originAddr" : "10.0.0.1", "application": "OpenSSH" , "protocol": "ssh" }`
)

var ts = httptest.NewServer(handlers())
var endpoint = fmt.Sprintf("%s%s", ts.URL, api.EventURL)

func init() {
	defaultEventClient = testEventClient
	log.SetLevel(log.WarnLevel)
	go runLookup()
}

func TestServerRequest(t *testing.T) {
	res, err := http.Post(fmt.Sprintf("%s%s", ts.URL, api.EventURL),
		"application/json",
		strings.NewReader(requestBody))

	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != http.StatusAccepted {
		t.Fatalf("Status code not 202 (%d)\n%s", res.StatusCode, string(b))
	}

	if res.ContentLength == 0 {
		t.Fatal("No Body")
	}

	var event Event
	if err != nil {
		t.Fatalf("Error reading body %s", err)
	}
	err = json.Unmarshal(b, &event)
	if err != nil {
		t.Fatalf("%s %s", string(b), err)
	}

	if event.ID == 0 {
		t.Fatalf("Event has id of 0 %+v", event)
	}

	time.Sleep(1 * time.Second)

	eventGeo := testEventClient.get(event.ID)
	if eventGeo == nil {
		t.Fatalf("Not not find id %d", event.ID)
	}

	if eventGeo.OriginCountry == "" {
		t.Fatal("Origin Country is null")
	}

	if eventGeo.RemoteCountry == "" {
		t.Fatal("Remote Country is null")
	}
}

func TestServerRequestWithOrigin(t *testing.T) {
	res, err := http.Post(endpoint,
		"application/json",
		strings.NewReader(requestBodyOrigin))

	if err != nil {
		t.Error(err)
	}

	var event Event
	b, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(b, &event)
	if err != nil {
		t.Fatalf("%s %s", string(b), err)
	}

	if event.ID == 0 {
		t.Fatalf("Event has id of 0 %+v", event)
	}

	time.Sleep(1 * time.Second)

	eventGeo := testEventClient.get(event.ID)
	if eventGeo == nil {
		t.Fatalf("Not not find id %d", event.ID)
	}

	if eventGeo.OriginCountry != "ZX" {
		t.Fatalf("Origin Country is not ZZ (%s)", eventGeo.OriginCountry)
	}

	if eventGeo.RemoteCountry == "" {
		t.Fatal("Remote Country is null")
	}
}

func BenchmarkServer(b *testing.B) {
	b.ReportAllocs()
	rw := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		handleEvent(rw, req(b))
		if rw.Result().StatusCode != 202 {
			b.Fatal("bad request")
		}
		reset(rw)
	}
}

func BenchmarkProcessEvent(b *testing.B) {
	var event Event
	b.ReportAllocs()
	if err := json.Unmarshal([]byte(requestBodyOrigin), &event); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		processEvent(event)
	}
}

func BenchmarkEvent(b *testing.B) {
	var event Event
	b.ReportAllocs()
	if err := json.Unmarshal([]byte(requestBodyOrigin), &event); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		defaultEventClient.recordEvent(event)
	}
}

func BenchmarkLookup(b *testing.B) {
	var event Event
	b.ReportAllocs()
	if err := json.Unmarshal([]byte(requestBodyOrigin), &event); err != nil {
		b.Fatal(err)
	}
	id, _, _ := defaultEventClient.recordEvent(event)
	event.ID = id
	for i := 0; i < b.N; i++ {
		defaultEventClient.resolveGeoEvent(event)
	}
}

func BenchmarkCache(b *testing.B) {
	var event Event
	clearDb(defaultEventClient.db, nil)
	b.ReportAllocs()
	if err := json.Unmarshal([]byte(requestBodyOrigin), &event); err != nil {
		b.Fatal(err)
	}
	id, _, _ := defaultEventClient.recordEvent(event)
	event.ID = id
	for i := 0; i < b.N; i++ {
		defaultEventClient.resolveGeoEvent(event)
	}
}

func req(t *testing.B) *http.Request {
	req, err := http.NewRequest("POST", api.EventURL, bufio.NewReader(strings.NewReader(requestBodyOrigin)))
	if err != nil {
		t.Fatal(err)
	}

	return req
}

func reset(rw *httptest.ResponseRecorder) {
	m := rw.HeaderMap
	for k := range m {
		delete(m, k)
	}
	body := rw.Body
	body.Reset()
	*rw = httptest.ResponseRecorder{
		Body:      body,
		HeaderMap: m,
	}

}
