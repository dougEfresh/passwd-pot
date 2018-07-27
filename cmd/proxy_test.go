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
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/dougEfresh/passwd-pot/api"
)

var request []byte

const (
	//requestBody       = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "1.2.3.4", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "application": "OpenSSH" , "protocol": "ssh"}`
	requestBodyOrigin = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "192.168.1.1", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "originAddr" : "10.0.0.1", "application": "OpenSSH" , "protocol": "ssh" }`
	//test_dsn          = "root@tcp(127.0.0.1:3306)/passwdpot?tls=false&parseTime=true&loc=UTC&timeout=50ms"
)

func init() {
	b := new(bytes.Buffer)
	b.WriteString(requestBodyOrigin)
	request = b.Bytes()
	proxyConfig.DryRun = true
}

func TestProxy(t *testing.T) {
	proxyConfig.bind = "localhost:8889"
	config.Debug = true
	go proxyrun("proxy")
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Post("http://localhost:8889", "application/octet-stream", bytes.NewReader(request))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatal("status code")
	}
	proxyRelayer.Drain()
	time.Sleep(350 * time.Millisecond)
	if len(proxyDryRunner.events) == 0 {
		t.Fatal("Relay not sent")
	}

	var event api.Event
	json.Unmarshal(request, &event)

	if event.User != proxyDryRunner.events[0].User {
		t.Fatal("!=user")
	}

	if event.Protocol != proxyDryRunner.events[0].Protocol {
		t.Fatal("!=")
	}

	if event.Passwd != proxyDryRunner.events[0].Passwd {
		t.Fatal("!=")
	}

	if event.Application != proxyDryRunner.events[0].Application {
		t.Fatal("!=")
	}

	if event.RemoteVersion != proxyDryRunner.events[0].RemoteVersion {
		t.Fatal("!=")
	}

	if event.RemotePort != proxyDryRunner.events[0].RemotePort {
		t.Fatal("!=")
	}

}

/* Fails
func BenchmarkproxyRelay_Send(b *testing.B) {
	b.ReportAllocs()
	proxyConfig.proxy = b.Name()
	go runproxyServer(mockSender{})
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < b.N; i++ {
		c, err := net.Dial("unix", b.Name())
		if err != nil {
			b.Fatal(err)
		}
		if _, err = c.Write(request); err != nil {
			b.Fatal(err)
		}
		c.Close()
	}
}
*/
