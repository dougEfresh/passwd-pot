package cmd

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUnMarshalEvent(t *testing.T) {
	var event SSHEvent
	err := json.Unmarshal([]byte(requestBodyOrigin), &event)
	if err != nil {
		t.Fatalf("Cannot deserialize event %s \n%s", err, requestBodyOrigin)
	}

	if event.Time.Time.UnixNano()/int64(time.Millisecond) != 1487973301661 {
		t.Fatalf("%s != 1487973301661", event.Time.Time)
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
	var event SSHEvent
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
			t.Fatalf("Error %d != 1487973301661", kv["time"])
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
