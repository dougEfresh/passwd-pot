package api

import (
	"github.com/Sirupsen/logrus"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var requestBody = `{"id": 1, "time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "192.168.1.1", "remotePort": 63185, "remoteName": "badguy.bad.com", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "originAddr" : "10.0.0.1", "application": "OpenSSH" , "protocol": "ssh" }`

var handler = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	logrus.Infof("Got request %s %s", r.Method, r.URL.Path)
	if r.Method == "POST" {
		io.WriteString(w, requestBody)
		w.WriteHeader(http.StatusAccepted)
	}
}

var server = httptest.NewServer(http.HandlerFunc(handler))

func TestSend(t *testing.T) {
	ec, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	ts := int64(1487973301661)
	event := &Event{
		ID:            0,
		Time:          EventTime(time.Unix(ts/1000, (ts%1000)*1000000).UTC()),
		User:          "admin",
		Passwd:        "12345678",
		RemoteAddr:    "192.168.1.1",
		RemotePort:    63185,
		RemoteName:    "badguy.bad.com",
		RemoteVersion: "SSH-2.0-JSCH-0.1.51",
		OriginAddr:    "",
		Application:   "OpenSSH",
		Protocol:      "ssh",
	}
	eventResp, err := ec.sendEvent(event)
	if err != nil {
		t.Fatalf("SendEvent: %s", err)
	}

	if eventResp == nil {
		t.Fatal("nill reponse")
	}

	if eventResp.ID != 1 {
		t.Fatalf("id should be 1 %s", eventResp)
	}
}
