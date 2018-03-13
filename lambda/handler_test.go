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

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

var body = ` { "originAddr": "not-an-address" , "event": { "time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "158.69.243.135", "originAddr": "212.143.121.156",  "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51", "application": "OpenSSH", "protocol": "ssh" } }`

func TestHandler(t *testing.T) {
	var e ApiEvent
	err := json.Unmarshal([]byte(body), &e)
	if err != nil {
		t.Fatalf("%s", err)
	}
	resp, err := Handle(e)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if len(resp.Body) <= 0 {
		t.Fatal("resp is crap")
	}
	if resp.StatusCode != 202 {
		t.Fatal("Not 202")
	}
	if !strings.Contains(resp.Body, "{id:") {
		t.Fatalf("%s", resp.Body)
	}
	t.Logf("Response is %s", resp.Body)
}
