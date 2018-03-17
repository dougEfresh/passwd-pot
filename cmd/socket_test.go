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
	"net"
	"testing"
	"time"
)

type mockSender struct {
}

func (ms mockSender) Send(r []byte) error {
	sent = r
	return nil
}

var request []byte
var sent []byte

const (
	requestBody       = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "1.2.3.4", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "application": "OpenSSH" , "protocol": "ssh"}`
	requestBodyOrigin = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "192.168.1.1", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "originAddr" : "10.0.0.1", "application": "OpenSSH" , "protocol": "ssh" }`
	//test_dsn          = "root@tcp(127.0.0.1:3306)/passwdpot?tls=false&parseTime=true&loc=UTC&timeout=50ms"
	test_dsn string = "postgres://postgres:@127.0.0.1/?sslmode=disable"
)

func init() {
	b := new(bytes.Buffer)
	b.WriteString("Host: localhost:8080\r\n")
	b.WriteString("User-Agent: curl/7.50.1\r\n")
	b.WriteString("Accept: */*\r\n")
	b.WriteString("Content-Length: 228\r\n\r\n")
	b.WriteString(requestBodyOrigin)
	request = b.Bytes()
}

func TestSocketRequest(t *testing.T) {
	socketConfig.Socket = t.Name()
	go runSocketServer(mockSender{})
	time.Sleep(500 * time.Millisecond)
	c, err := net.Dial("unix", t.Name())

	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, err = c.Write(request)

	if err != nil {
		t.Fatal(err)
	}
	var resp []byte = make([]byte, 128)
	if _, err = c.Read(resp); err != nil {
		t.Fatal(err)
	}
	time.Sleep(250 * time.Millisecond)
	if sent == nil || len(sent) == 0 {
		t.Fatal("Relay not setn")
	}

	if !bytes.Equal(sent, request) {
		t.Fatalf("Not the same %s %s", string(sent), string(request))
	}
}

/* Fails
func BenchmarkSocketRelay_Send(b *testing.B) {
	b.ReportAllocs()
	socketConfig.Socket = b.Name()
	go runSocketServer(mockSender{})
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
