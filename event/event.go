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

package event

import (
	"database/sql"
	"os"
	"strings"

	"github.com/dougEfresh/passwd-pot/potdb"

	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
)

type EventClient struct {
	db     potdb.DB
	logger log.FieldLogger
}

type EventOptionFunc func(*EventClient) error

func NewEventClient(options ...EventOptionFunc) (*EventClient, error) {
	ec := &EventClient{}
	for _, option := range options {
		if err := option(ec); err != nil {
			return nil, err
		}
	}
	if ec.logger == nil {
		ec.logger = defaultLogger
	}
	return ec, nil
}

func SetEventDb(db potdb.DB) EventOptionFunc {
	return func(c *EventClient) error {
		c.db = db
		return nil
	}
}

func WithDsn(dsn string) EventOptionFunc {
	return func(c *EventClient) error {
		var err error
		c.db, err = potdb.Open(dsn)
		return err
	}
}

func SetEventLogger(l log.FieldLogger) EventOptionFunc {
	return func(c *EventClient) error {
		c.logger = l
		return nil
	}
}

func (c *EventClient) RecordEvent(event api.Event) (int64, error) {
	result, err := c.db.Insert(`INSERT INTO event
	(dt, username, passwd, remote_addr, remote_port, remote_name, remote_version, origin_addr, application, protocol)
        VALUES(?,?,?,?,?,?,?,?,?,?)
        `, event.Time, event.User, event.Passwd, event.RemoteAddr, event.RemotePort, event.RemoteName, event.RemoteVersion, event.OriginAddr, event.Application, event.Protocol)
	if err == nil {
		return result.LastInsertId()
	}
	return 0, err
}

func (c *EventClient) GetEvent(id int64) (*api.EventGeo, error) {
	r := c.db.QueryRow(`SELECT
	id, dt, username, passwd, remote_addr, remote_name, remote_version, remote_port, remote_country, remote_city,
	origin_addr, origin_country, origin_city,
	remote_latitude, remote_longitude,
        origin_latitude, origin_longitude
	FROM event_geo WHERE id = ? LIMIT 1`, id)
	var event api.EventGeo
	err := r.Scan(&event.ID, &event.Time, &event.User, &event.Passwd,
		&event.RemoteAddr, &event.RemoteName, &event.RemoteVersion,
		&event.RemotePort, &event.RemoteCountry, &event.RemoteCity,
		&event.OriginAddr, &event.OriginCountry, &event.OriginCity,
		&event.RemoteLatitude, &event.RemoteLongitude,
		&event.OriginLatitude, &event.OriginLongitude)
	if err != nil {
		c.logger.Errorf("Error getting event id %d %s", id, err)
		return nil, err
	}
	return &event, nil
}

func (c *EventClient) GetCountryStats() ([]api.CountryStat, error) {
	r, err := c.db.Query(`SELECT country_code,sum(hits) as hits FROM country_stats GROUP BY country_code ORDER BY country_code`)
	var stats = make([]api.CountryStat, 5000)
	var cnt = 0
	if err != nil {
		return nil, err
	}
	for r.Next() {
		var stat api.CountryStat
		r.Scan(&stat.Country, &stat.Hits)
		if len(stats) > cnt+1 {
			var buf = make([]api.CountryStat, 1000)
			stats = append(stats, buf[0:999]...)
		}
		stats[cnt] = stat
		cnt++
	}
	return stats[0:cnt], nil
}

var defaultLogger log.FieldLogger

func init() {
	defaultLogger = log.DefaultLogger(os.Stdout)
}

func loadDSN(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	if strings.Contains(dsn, "postgres") {
		db, err = sql.Open("postgres", dsn)
	} else {
		db, err = sql.Open("mysql", dsn)
	}
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}