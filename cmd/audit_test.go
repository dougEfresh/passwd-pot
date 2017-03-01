package cmd

import (
	"github.com/dougEfresh/dbr"
	"testing"
	"time"
	"encoding/json"
)

var localGeo = make(map[string]string)

func init() {
	localGeo["1.2.3.4"] = `{"ip":"1.2.3.4","country_code":"CA","country_name":"Singapore","region_code":"01","region_name":"Central Singapore Community Development Council","city":"Singapore","zip_code":"","time_zone":"Asia/Singapore","latitude":1.1,"longitude":101.00,"metro_code":0}`
	localGeo["127.0.0.1"] = `{"ip":"127.0.0.1","country_code":"US","country_name":"USA","region_code":"05","region_name":"America","city":"New York","zip_code":"","time_zone":"Asia/Singapore","latitude":2.2,"longitude":102.00,"metro_code":0}`
}

func (c *mockGeoClient) GetLocationForIP(ip string) (*Geo, error) {
	resp := []byte(localGeo[ip])
	var geo = &Geo{}
	err := json.Unmarshal(resp, geo)
	geo.Ip = ip
	geo.LastUpdate = time.Now()
	return geo, err
}

const dsn string = "postgres://ssh_audit:ssh_audit@localhost/ssh_audit"

var testAuditClient = &AuditClient{
	db:        loadDSN(dsn),
	geoClient: GeoClientTransporter(&mockGeoClient{}),
}

func clearDb(db *dbr.Connection, t *testing.T) {
	sess := db.NewSession(nil)
	if _, err := sess.DeleteFrom("event").Exec(); err != nil {
		t.Fatalf("Error deletiing %s", err)
	}
	if _, err := sess.DeleteFrom("geo").Exec(); err != nil {
		t.Fatalf("Error deletiing %s", err)
	}
}

var now = time.Now()
var event = SshEvent{
	RemoteAddr:    "1.2.3.4",
	RemotePort:    3432,
	RemoteVersion: "SSH-2.0-JSCH-0.1.51",
	RemoteName:    "blah",
	User:          "admin",
	Passwd:        "1234",
	Time:         JsonTime(now),
	OriginAddr:    "127.0.0.1",
}

func createEvent(event *SshEvent) error {
	sess := testAuditClient.db.NewSession(nil)
	err := testAuditClient.RecordEvent(event)
	if err != nil {
		return err
	}
	var eventGeo SshEventGeo
	_, err = sess.Select("*").From("vw_event").Where("id = ?", event.ID).Load(&eventGeo)
	if err != nil {
		return err
	}
	return nil
}

func TestRecordEvent(t *testing.T) {
	clearDb(testAuditClient.db, t)
	err := createEvent(&event)
	if err != nil {
		t.Fatalf("Error creting event %s", err)
	}

	if event.ID <= 0 {
		t.Fatalf("Event id should be > 0 %+v", &event)
	}

}

func TestLookup(t *testing.T) {
	TestRecordEvent(t)
	err := createEvent(&event)
	if err != nil {
		t.Fatalf("Error creting event %s", err )
	}

	if event.ID <= 0 {
		t.Fatalf("Event id should be > 0 %+v", &event)
	}

	testAuditClient.ResolveGeoEvent(&event)
	geoEvent := testAuditClient.Get(event.ID)

	if geoEvent == nil {
		t.Fatalf("Could not find id %d", event.ID)
	}

	if geoEvent.RemoteAddr != event.RemoteAddr {
		t.Fatalf("%s != %s", geoEvent.RemoteAddr, event.RemoteAddr)
	}

	if geoEvent.RemotePort != event.RemotePort {
		t.Fatalf("%d != %d", geoEvent.RemotePort, event.RemotePort)
	}

	if geoEvent.RemoteName != event.RemoteName {
		t.Fatalf("%s != %s", geoEvent.RemoteName, event.RemoteName)
	}

	if geoEvent.RemoteVersion != event.RemoteVersion {
		t.Fatalf("%s != %s", geoEvent.RemoteVersion, event.RemoteVersion)
	}

	if geoEvent.User != event.User {
		t.Fatalf("%s != %s", geoEvent.User, event.User)
	}

	if geoEvent.Passwd != event.Passwd {
		t.Fatalf("%s != %s", geoEvent.Passwd, event.Passwd)
	}

	if geoEvent.RemoteCountry != "CA" {
		t.Fatalf("%s != CA", geoEvent.RemoteCountry)
	}

	if geoEvent.RemoteCity != "Singapore" {
		t.Fatalf("%s != Singapore", geoEvent.RemoteCountry)
	}

	if geoEvent.OriginAddr != "127.0.0.1" {
		t.Fatalf("%s != 127.0.0.1", geoEvent.OriginAddr)
	}

	if geoEvent.OriginCountry != "US" {
		t.Fatalf("%s != US", geoEvent.OriginCountry)
	}

	if geoEvent.OriginCity != "New York" {
		t.Fatalf("%s != New York", geoEvent.OriginCity)
	}

	if geoEvent.RemoteLatitude != 1.1 {
		t.Fatalf("%s != 1.1", geoEvent.RemoteLatitude)
	}

	if geoEvent.RemoteLongitude != 101.00 {
		t.Fatalf("%s != 101.00", geoEvent.RemoteLongitude)
	}

	if geoEvent.OriginLatitude != 2.2 {
		t.Fatalf("%s != 2.2", geoEvent.OriginLatitude)
	}

	if geoEvent.OriginLongitude != 102.00 {
		t.Fatalf("%s != 102.00", geoEvent.OriginLongitude)
	}
}
