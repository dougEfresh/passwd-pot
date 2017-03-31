package cmd

import (
	"database/sql"
	"encoding/json"
	"github.com/dougEfresh/passwd-pot/api"
	"testing"
	"time"
)

var localGeo = make(map[string]string)

type mockGeoClient struct {
}

func init() {
	localGeo["1.2.3.4"] = `{"ip":"1.2.3.4","country_code":"CA","country_name":"Singapore","region_code":"01","region_name":"Central Singapore Community Development Council","city":"Singapore","zip_code":"","time_zone":"Asia/Singapore","latitude":1.1,"longitude":101.00,"metro_code":0}`
	localGeo["127.0.0.1"] = `{"ip":"127.0.0.1","country_code":"US","country_name":"USA","region_code":"05","region_name":"America","city":"New York","zip_code":"","time_zone":"Asia/Singapore","latitude":2.2,"longitude":102.00,"metro_code":0}`
	localGeo["192.168.1.1"] = `{"ip":"192.168.1.1","country_code":"ZZ","country_name":"USA","region_code":"05","region_name":"America","city":"New York","zip_code":"","time_zone":"Asia/Singapore","latitude":2.2,"longitude":102.00,"metro_code":0}`
	localGeo["10.0.0.1"] = `{"ip":"10.0.0.1","country_code":"ZX","country_name":"USA","region_code":"05","region_name":"America","city":"New York","zip_code":"","time_zone":"Asia/Singapore","latitude":2.2,"longitude":102.00,"metro_code":0}`
}

func (c *mockGeoClient) getLocationForAddr(ip string) (*Geo, error) {
	resp := []byte(localGeo[ip])
	var geo = &Geo{}
	err := json.Unmarshal(resp, geo)
	geo.IP = ip
	geo.LastUpdate = time.Now()
	return geo, err
}

const test_dsn string = "postgres://postgres:@127.0.0.1:5432/?sslmode=disable"

var testEventClient = &eventClient{
	db:        loadDSN(test_dsn),
	geoClient: geoClientTransporter(&mockGeoClient{}),
}

func clearDb(db *sql.DB, t *testing.T) {
	for k, _ := range localGeo {
		geoCache.Delete(k)
	}
	if _, err := db.Exec("DELETE FROM event"); err != nil {
		t.Fatalf("Error deletiing %s", err)
	}
	if _, err := db.Exec("DELETE FROM geo"); err != nil {
		t.Fatalf("Error deletiing %s", err)
	}
}

var now = time.Now()
var testEvent = Event{
	RemoteAddr:    "1.2.3.4",
	RemotePort:    3432,
	RemoteVersion: "SSH-2.0-JSCH-0.1.51",
	RemoteName:    "blah",
	User:          "admin",
	Passwd:        "1234",
	Time:          api.EventTime(now),
	OriginAddr:    "127.0.0.1",
	Application:   "OpenSSH",
	Protocol:      "ssh",
}

func createEvent(event *Event) error {
	id, _, err := testEventClient.recordEvent(*event)
	if err != nil {
		return err
	}
	r := testEventClient.db.QueryRow(`SELECT id FROM event_geo WHERE id = $1 LIMIT 1`, id)
	err = r.Scan(&id)
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

func TestRecordEvent(t *testing.T) {
	clearDb(testEventClient.db, t)
	err := createEvent(&testEvent)
	if err != nil {
		t.Fatalf("Error creting event %s", err)
	}

	if testEvent.ID <= 0 {
		t.Fatalf("Event id should be > 0 %+v", &testEvent)
	}

}

func TestLookup(t *testing.T) {
	clearDb(testEventClient.db, t)
	err := createEvent(&testEvent)
	if err != nil {
		t.Fatalf("Error creting event %s", err)
	}

	if testEvent.ID <= 0 {
		t.Fatalf("Event id should be > 0 %+v", &testEvent)
	}

	err = testEventClient.resolveGeoEvent(testEvent)
	if err != nil {
		t.Fatalf("Error with getting geo %s", err)
	}
	geoEvent := testEventClient.get(testEvent.ID)

	if geoEvent == nil {
		t.Fatalf("Could not find id %d", testEvent.ID)
	}

	if geoEvent.RemoteAddr != testEvent.RemoteAddr {
		t.Fatalf("%s != %s", geoEvent.RemoteAddr, testEvent.RemoteAddr)
	}

	if geoEvent.RemotePort != testEvent.RemotePort {
		t.Fatalf("%d != %d", geoEvent.RemotePort, testEvent.RemotePort)
	}

	if geoEvent.RemoteName != testEvent.RemoteName {
		t.Fatalf("%s != %s", geoEvent.RemoteName, testEvent.RemoteName)
	}

	if geoEvent.RemoteVersion != testEvent.RemoteVersion {
		t.Fatalf("%s != %s", geoEvent.RemoteVersion, testEvent.RemoteVersion)
	}

	if geoEvent.User != testEvent.User {
		t.Fatalf("%s != %s", geoEvent.User, testEvent.User)
	}

	if geoEvent.Passwd != testEvent.Passwd {
		t.Fatalf("%s != %s", geoEvent.Passwd, testEvent.Passwd)
	}

	if geoEvent.RemoteCountry != "CA" {
		t.Fatalf("%s != CA (%s)", geoEvent.RemoteCountry, geoEvent)
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
		t.Fatalf("%f != 1.1", geoEvent.RemoteLatitude)
	}

	if geoEvent.RemoteLongitude != 101.00 {
		t.Fatalf("%f != 101.00", geoEvent.RemoteLongitude)
	}

	if geoEvent.OriginLatitude != 2.2 {
		t.Fatalf("%f != 2.2", geoEvent.OriginLatitude)
	}

	if geoEvent.OriginLongitude != 102.00 {
		t.Fatalf("%f != 102.00", geoEvent.OriginLongitude)
	}
}

func TestExpire(t *testing.T) {
	clearDb(testEventClient.db, t)
	err := createEvent(&testEvent)

	if err != nil {
		t.Fatalf("Error creting event %s", err)
	}

	if testEvent.ID <= 0 {
		t.Fatalf("Event id should be > 0 %+v", &testEvent)
	}

	testEventClient.resolveGeoEvent(testEvent)
	geoEvent := testEventClient.get(testEvent.ID)

	if geoEvent == nil {
		t.Fatalf("Could not find id %d", testEvent.ID)
	}
	geoCache.Delete(testEvent.RemoteAddr)
	var oldlastUpdate time.Time
	var newerLastUpdate time.Time
	r := testEventClient.db.QueryRow("select last_update from geo where ip = $1 LIMIT 1", testEvent.RemoteAddr)
	err = r.Scan(&oldlastUpdate)
	if err != nil {
		t.Fatalf("Error updating time %s", err)
	}
	_, err = testEventClient.db.Exec("UPDATE geo SET last_update = $1", time.Now().Add(time.Hour*24*-100))
	if err != nil {
		t.Fatalf("Error updating time %s", err)
	}
	createEvent(&testEvent)
	testEventClient.resolveGeoEvent(testEvent)
	r = testEventClient.db.QueryRow("select last_update from geo where ip = $1 LIMIT 1", testEvent.RemoteAddr)
	err = r.Scan(&newerLastUpdate)

	if err != nil {
		t.Fatalf("Error updating time %s", err)
	}
	if oldlastUpdate.After(newerLastUpdate) {
		t.Fatalf("old is afer new %s > %s", oldlastUpdate, newerLastUpdate)
	}
}

func TestExpireAndChangedGeo(t *testing.T) {
	clearDb(testEventClient.db, t)
	err := createEvent(&testEvent)

	if err != nil {
		t.Fatalf("Error creting event %s", err)
	}

	if testEvent.ID <= 0 {
		t.Fatalf("Event id should be > 0 %+v", &testEvent)
	}

	testEventClient.resolveGeoEvent(testEvent)
	geoEvent := testEventClient.get(testEvent.ID)

	if geoEvent == nil {
		t.Fatalf("Could not find id %d", testEvent.ID)
	}
	var oldlastUpdate time.Time
	var newerLastUpdate time.Time
	r := testEventClient.db.QueryRow("SELECT last_update FROM geo WHERE ip = $1 LIMIT 1", testEvent.RemoteAddr)
	err = r.Scan(&oldlastUpdate)
	if err != nil {
		t.Fatalf("Error updating time %s", err)
	}

	geoCache.Delete(testEvent.RemoteAddr)
	_, err = testEventClient.db.Exec("UPDATE geo SET last_update = $1 WHERE ip = $2", time.Now().Add(time.Hour*24*-100), testEvent.RemoteAddr)
	if err != nil {
		t.Fatalf("Error updating time %s", err)
	}
	oldGeo := localGeo["1.2.3.4"]
	localGeo["1.2.3.4"] = `{"ip":"1.2.3.4","country_code":"DE","country_name":"Singapore","region_code":"01","region_name":"Central Singapore Community Development Council","city":"Singapore","zip_code":"","time_zone":"Asia/Singapore","latitude":1.1,"longitude":101.00,"metro_code":0}`
	//reset to original geo
	defer func() {
		localGeo["1.2.3.4"] = oldGeo
	}()

	createEvent(&testEvent)
	testEventClient.resolveGeoEvent(testEvent)
	geoEvent = testEventClient.get(testEvent.ID)
	if geoEvent == nil {
		t.Fatalf("Could not find id %d", testEvent.ID)
	}

	r = testEventClient.db.QueryRow("SELECT last_update FROM geo WHERE ip = $1 ORDER BY last_update DESC LIMIT 1", testEvent.RemoteAddr)
	err = r.Scan(&newerLastUpdate)
	if err != nil {
		t.Fatalf("Error getting  %s", err)
	}
	if oldlastUpdate.After(newerLastUpdate) {
		t.Fatalf("old is afer new %s > %s", oldlastUpdate, newerLastUpdate)
	}
	if geoEvent.RemoteCountry != "DE" {
		t.Fatalf("Country code not DE (%s)", geoEvent.RemoteCountry)
	}
}
