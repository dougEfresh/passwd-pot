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
	"testing"
	"time"
)

func TestUnMarshalEvent(t *testing.T) {
	var event Event
	err := json.Unmarshal([]byte(requestBodyOrigin), &event)
	if err != nil {
		t.Fatalf("Cannot deserialize event %s \n%s", err, requestBodyOrigin)
	}

	if time.Time(event.Time).UnixNano()/int64(time.Millisecond) != 1487973301661 {
		t.Fatalf("%s != 1487973301661", event.Time.String())
	}

	if event.OriginAddr != "10.0.0.1" {
		t.Fatalf("OriginAddr != 10.0.0.1 (%s)", event.OriginAddr)
	}

	if event.OriginAddr != "10.0.0.1" {
		t.Fatalf("OriginAddr != 10.0.0.1 (%s)", event.OriginAddr)
	}

	if event.RemoteAddr != "192.168.1.1" {
		t.Fatalf("OriginAddr != 192.168.1.1 (%s)", event.RemoteAddr)
	}
}

//Verify time is converted properly
func TestMarshalEvent(t *testing.T) {
	var event Event
	var kv map[string]interface{}
	err := json.Unmarshal([]byte(requestBodyOrigin), &event)
	b, err := json.Marshal(&event)
	if err != nil {
		t.Fatalf("Cannot deserialize event %s \n%s", err, requestBodyOrigin)
	}

	err = json.Unmarshal(b, &kv)
	if err != nil {
		t.Fatalf("Error %s", err)
	}

	switch i := kv["time"].(type) {
	case float64:
		if i != float64(1487973301661) {
			t.Fatalf("Error %s != 1487973301661", kv["time"])
		}
	default:
		t.Fatal("Unknown type")
	}
}

func TestEventEquals(t *testing.T) {
	g := Geo{
		ID:          0,
		IP:          testEvent.RemoteAddr,
		CountryCode: "US",
		City:        "Mexico",
		RegionName:  "A galaxy far far away",
		RegionCode:  "666",
		TimeZone:    "GMT",
		Latitude:    1,
		Longitude:   2,
		MetroCode:   69,
	}

	another := Geo{
		ID:          0,
		IP:          testEvent.RemoteAddr,
		CountryCode: "US",
		City:        "New York",
		RegionName:  "A galaxy far far away",
		RegionCode:  "777",
		TimeZone:    "GMT",
		Latitude:    1,
		Longitude:   2,
		MetroCode:   69,
	}

	notEqual := Geo{
		ID:          0,
		IP:          testEvent.RemoteAddr,
		CountryCode: "CA",
		City:        "New York",
		RegionName:  "A galaxy far far away",
		RegionCode:  "777",
		TimeZone:    "GMT",
		Latitude:    1,
		Longitude:   2,
		MetroCode:   69,
	}

	if !g.equals(&g) {
		t.Fatalf("%+v != %+v", &g, &g)

	}

	if g.equals(&another) {
		t.Fatalf("%+v == %+v", &g, &another)

	}

	if g.equals(&notEqual) {
		t.Fatalf("%+v == %+v", &g, &notEqual)
	}
}
