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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var requestBody = `{"id": 1, "time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "192.168.1.1", "remotePort": 63185, "remoteName": "badguy.bad.com", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "originAddr" : "10.0.0.1", "application": "OpenSSH" , "protocol": "ssh" }`

var handler = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if r.Method == "POST" {
		io.WriteString(w, requestBody)
		w.WriteHeader(http.StatusAccepted)
	}
}

var server = httptest.NewServer(http.HandlerFunc(handler))

func TestSend(t *testing.T) {
	ec, err := New(server.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	ts := int64(1487973301661)
	event := Event{
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
	_, err = ec.RecordEvent(event)
	if err != nil {
		t.Fatalf("RecordEvent: %s", err)
	}
}
