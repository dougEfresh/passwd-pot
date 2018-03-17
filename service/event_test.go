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

package service

import (
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
	klog "github.com/go-kit/kit/log"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var testEventClient = &EventClient{}

//const test_dsn = "root@tcp(127.0.0.1:3306)/passwdpot?tls=false&parseTime=true&loc=UTC&timeout=50ms"
const testdsn string = "postgres://postgres:@127.0.0.1/?sslmode=disable"

func init() {
	dsn := os.Getenv("PASSWDPOT_DSN")
	var db *sql.DB
	if dsn == "" {
		db, _ = loadDSN(testdsn)
	} else {
		db, _ = loadDSN(dsn)
	}
	testEventClient, _ = NewEventClient(SetEventDb(db))
	testEventClient.mysql = !strings.Contains(testdsn, "postgres")
	defaultLogger.AddLogger(klog.NewJSONLogger(os.Stderr))
	defaultLogger.SetLevel(log.WarnLevel)
}

func clearDb(db *sql.DB, t *testing.T) {
	if _, err := db.Exec("DELETE FROM event"); err != nil {
		t.Fatalf("Error deletiing %s", err)
	}
	if _, err := db.Exec("DELETE FROM geo"); err != nil {
		t.Fatalf("Error deletiing %s", err)
	}
}

var now = time.Now()
var testEvent = api.Event{
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

func createEvent(event *api.Event) error {
	id, err := testEventClient.RecordEvent(*event)
	if err != nil {
		return err
	}
	r := testEventClient.db.QueryRow(testEventClient.replaceParams(`SELECT id FROM event_geo WHERE id = ? LIMIT 1`), id)
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

/*func TestEventClient_GetCountryStats(t *testing.T) {
	testEventClient.db.Query("DELETE FROM country_stats")
	testEventClient.db.Query("INSERT INTO country_stats VALUES ('US',1.0,2.0,1234)")
	testEventClient.db.Query("INSERT INTO country_stats VALUES ('CA',3.0,4.0,56789)")
	stats, err := testEventClient.GetCountryStats()
	if err != nil {
		t.Fatalf("Error getting stats %s", err)
	}
	if len(stats) != 2 {
		t.Fatalf("Stats != 2 (%d)", len(stats))
	}
	var s = stats[1]
	if s.Country != "US" {
		t.Fatalf("!= US '%s'", s.Country)
	}
	if s.Latitude != 0.0 {
		t.Fatalf("!= 0.0")
	}

	if s.Longitude != 0.0 {
		t.Fatalf("!= 0.0")
	}

	if s.Hits != 1234 {
		t.Fatalf("!= 12345")
	}

	testEventClient.db.Query("INSERT INTO country_stats VALUES ('CH',5.0,6.0,56789)")
	stats, err = testEventClient.GetCountryStats()
	if err != nil {
		t.Fatalf("Error getting stats %s", err)
	}
	if len(stats) != 3 {
		t.Fatalf("Stats != 3 (%d)", len(stats))
	}
}
*/
