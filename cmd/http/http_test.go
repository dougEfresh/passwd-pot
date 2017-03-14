package http

import (
	"encoding/base64"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dougEfresh/passwd-pot/api"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var submittedEvent *api.Event

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

type BasicAuthTransport struct {
	Username string
	Password string
}

func (bat BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
			bat.Username, bat.Password)))))
	return http.DefaultTransport.RoundTrip(req)
}

func (bat *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: bat}
}

type mockQueue struct {
}

func (mq *mockQueue) Send(e *api.Event) {
	submittedEvent = e
}

func TestServerRequest(t *testing.T) {
	var ts = httptest.NewServer(&potHttpHandler{
		eventQueue: &mockQueue{},
	})
	res, err := http.Post(fmt.Sprintf("%s%s", ts.URL, api.EventURL),
		"application/json",
		strings.NewReader(""))

	if err != nil {
		t.Fatalf("Error! %s", err)
	}

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Error! status should be 401 %d", res.StatusCode)
	}
	res.Body.Close()
	ba := &BasicAuthTransport{
		Username: "blah",
		Password: "me",
	}

	req, err := http.NewRequest("GET", ts.URL+"/something", nil)
	res, err = ba.Client().Do(req)
	if err != nil {
		t.Fatalf("Could not do basic auth %s", err)
	}
	time.Sleep(1 * time.Second)
	if submittedEvent == nil {
		t.Fatal("Submitted event is null")
	}
	if submittedEvent.User != "blah" {
		t.Fatalf("Wrong event sent %s", submittedEvent)
	}

	if submittedEvent.Passwd != "me" {
		t.Fatalf("Wrong event sent %s", submittedEvent)
	}

	if !strings.Contains(submittedEvent.RemoteVersion, "Go") {
		t.Fatalf("Wrong event sent %s", submittedEvent)
	}

	if !strings.Contains(submittedEvent.RemoteAddr, "127.0.0.1") {
		t.Fatalf("Wrong event sent %s", submittedEvent)
	}

	if !strings.Contains(submittedEvent.Protocol, "http-basic-auth") {
		t.Fatalf("Wrong event sent %s", submittedEvent)
	}

	if submittedEvent.RemotePort == 0 {
		t.Fatalf("Wrong event sent %s", submittedEvent)
	}

	if submittedEvent.RemoteName == "" {
		t.Fatalf("Wrong event sent %s", submittedEvent)
	}

	defer res.Body.Close()
}

func BenchmarkRunHttpPot(b *testing.B) {
	b.ReportAllocs()
	var ts = httptest.NewServer(&potHttpHandler{
		eventQueue: &mockQueue{},
	})
	logrus.SetLevel(logrus.InfoLevel)
	defer logrus.SetLevel(logrus.DebugLevel)
	for i := 0; i < b.N; i++ {
		res, err := http.Post(fmt.Sprintf("%s%s", ts.URL, api.EventURL),
			"application/json",
			strings.NewReader(""))
		if err != nil {
			b.Fatalf("Error! %s", err)
		}
		res.Body.Close()
	}
}
