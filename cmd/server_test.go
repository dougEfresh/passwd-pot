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
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
	"github.com/dougEfresh/passwd-pot/service"
	klog "github.com/go-kit/kit/log"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	requestBody       = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "1.2.3.4", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "application": "OpenSSH" , "protocol": "ssh"}`
	requestBodyOrigin = `{"time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "192.168.1.1", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51" , "originAddr" : "10.0.0.1", "application": "OpenSSH" , "protocol": "ssh" }`
	test_dsn          = "root@tcp(127.0.0.1:3306)/passwdpot?tls=false&parseTime=true&loc=UTC&timeout=50ms"
)

var ts *httptest.Server
var eventEndpoint string
var localGeo = make(map[string]string)

type mockGeoClient struct {
}

func init() {
	localGeo["1.2.3.4"] = `{"ip":"1.2.3.4","country_code":"CA","country_name":"Singapore","region_code":"01","region_name":"Central Singapore Community Development Council","city":"Singapore","zip_code":"","time_zone":"Asia/Singapore","latitude":1.1,"longitude":101.00,"metro_code":0}`
	localGeo["127.0.0.1"] = `{"ip":"127.0.0.1","country_code":"US","country_name":"USA","region_code":"05","region_name":"America","city":"New York","zip_code":"","time_zone":"Asia/Singapore","latitude":2.2,"longitude":102.00,"metro_code":0}`
	localGeo["192.168.1.1"] = `{"ip":"192.168.1.1","country_code":"ZZ","country_name":"USA","region_code":"05","region_name":"America","city":"New York","zip_code":"","time_zone":"Asia/Singapore","latitude":2.2,"longitude":102.00,"metro_code":0}`
	localGeo["10.0.0.1"] = `{"ip":"10.0.0.1","country_code":"ZX","country_name":"USA","region_code":"05","region_name":"America","city":"New York","zip_code":"","time_zone":"Asia/Singapore","latitude":2.2,"longitude":102.00,"metro_code":0}`
}

func (c *mockGeoClient) GetLocationForAddr(ip string) (*service.Geo, error) {
	resp := []byte(localGeo[ip])
	logger.Infof("Found mocked %s", localGeo[ip])
	var geo = &service.Geo{}
	err := json.Unmarshal(resp, geo)
	geo.IP = ip
	geo.LastUpdate = time.Now()
	return geo, err
}

func init() {
	dsn := os.Getenv("PASSWDPOT_DSN")
	var db *sql.DB
	if dsn == "" {
		db = loadDSN(test_dsn)
	} else {
		db = loadDSN(dsn)
	}

	logger.SetLevel(log.DebugLevel)
	logger.AddLogger(klog.NewJSONLogger(os.Stdout))
	logger.With("ts", klog.DefaultTimestamp)
	logger.With("caller", klog.DefaultCaller)
	eventClient, _ = service.NewEventClient(service.SetEventDb(db), service.SetEventLogger(logger))
	resolveClient, _ = service.NewResolveClient(service.SetResolveDb(db), service.SetResolveLogger(logger), service.SetGeoClient(service.GeoClientTransporter(&mockGeoClient{})))
	ts = httptest.NewServer(handlers())
	eventEndpoint = fmt.Sprintf("%s%s", ts.URL, api.EventURL)

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
	eventGeo, err := eventClient.GetEvent(id)
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
	eventGeo, _ := eventClient.GetEvent(id)

	if eventGeo == nil {
		t.Fatalf("Not not find id %d", id)
	}

	if eventGeo.OriginCountry != "ZX" {
		t.Fatalf("Origin Country is not ZX (%s)", eventGeo.OriginCountry)
	}

	if eventGeo.RemoteCountry == "" {
		t.Fatal("Remote Country is null")
	}
}
