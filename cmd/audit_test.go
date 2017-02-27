package cmd

import (
	"github.com/dougEfresh/dbr"
	"testing"
	"time"
)

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

var now = time.Now().UnixNano() / 1000000
var event = SshEvent{
	RemoteAddr:    "1.2.3.4",
	RemotePort:    3432,
	RemoteVersion: "SSH-2.0-JSCH-0.1.51",
	RemoteName:    "blah",
	User:          "admin",
	Passwd:        "1234",
	Epoch:         now,
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
		t.Fatalf("Error creting event %s", err )
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
}
