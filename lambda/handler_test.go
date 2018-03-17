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
	"testing"

	"github.com/dougEfresh/passwd-pot/api"
)

var body = `{"originAddr": "127.0.0.1", "time": 1148797330161, "user": "admin", "passwd": "12345678", "remoteAddr": "4.2.2.2", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-", "application": "OpenSSH", "protocol": "ssh"}`

func TestHandler(t *testing.T) {
	var e api.Event
	json.Unmarshal([]byte(body), e)
	resp, err := Handle(e)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if resp.ID <= 0 {
		t.Fatal("resp is crap")
	}

	if id, found := geoCache.Get("4.2.2.2"); !found {
		t.Fatalf("Cannot find 4.2.2.2 ip in cache (%d)", id)
	}

	t.Logf("Response is %d", resp.ID)
}

func TestHandlerError(t *testing.T) {
	e := api.Event{}
	_, err := Handle(e)
	if err == nil {
		t.Fatal("There should be an error")
	}
}
