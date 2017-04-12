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
	"encoding/json"
	"fmt"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/log"
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

var ts *httptest.Server
var eventEndpoint string

func init() {
	h, _ := getHandler(testEventClient)
	ts = httptest.NewServer(h)
	eventEndpoint = fmt.Sprintf("%s%s", ts.URL, api.EventURL)
	logger.SetLevel(log.WarnLevel)
}

func TestServerRequest(t *testing.T) {
	res, err := http.Post(eventEndpoint,
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

	var id int64
	if err != nil {
		t.Fatalf("Error reading body %s", err)
	}
	err = json.Unmarshal(b, &id)
	if err != nil {
		t.Fatalf("%s %s", string(b), err)
	}
	time.Sleep(1 * time.Second)
	eventGeo := testEventClient.get(id)
	if eventGeo == nil {
		t.Fatalf("Not not find id %d", id)
	}
	if eventGeo.OriginCountry == "" {
		t.Fatal("Origin Country is null")
	}
	if eventGeo.RemoteCountry == "" {
		t.Fatal("Remote Country is null")
	}
}

func TestServerRequestWithOrigin(t *testing.T) {
	res, err := http.Post(eventEndpoint,
		"application/json",
		strings.NewReader(requestBodyOrigin))

	if err != nil {
		t.Error(err)
	}

	b, _ := ioutil.ReadAll(res.Body)
	var id int64
	if err != nil {
		t.Fatalf("Error reading body %s", err)
	}
	err = json.Unmarshal(b, &id)
	if err != nil {
		t.Fatalf("%s %s", string(b), err)
	}

	time.Sleep(500 * time.Millisecond)
	eventGeo := testEventClient.get(id)
	if eventGeo == nil {
		t.Fatalf("Not not find id %d", id)
	}

	if eventGeo.OriginCountry != "ZX" {
		t.Fatalf("Origin Country is not ZZ (%s)", eventGeo.OriginCountry)
	}

	if eventGeo.RemoteCountry == "" {
		t.Fatal("Remote Country is null")
	}
}
