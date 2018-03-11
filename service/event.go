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
	"fmt"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

type EventClient struct {
	db     *sql.DB
	logger log.Logger
	mysql  bool
}

type EventOptionFunc func(*EventClient) error

func NewEventClient(options ...EventOptionFunc) (*EventClient, error) {
	ec := &EventClient{
		logger: defaultLogger,
	}
	for _, option := range options {
		if err := option(ec); err != nil {
			return nil, err
		}
	}
	return ec, nil
}

func SetEventDb(db *sql.DB) EventOptionFunc {
	return func(c *EventClient) error {
		c.db = db
		return nil
	}
}

func WithDsn(dsn string) EventOptionFunc {
	return func(c *EventClient) error {
		var err error
		c.db, err = loadDSN(dsn)
		c.mysql = !strings.Contains(dsn, "postgres")
		return err
	}
}

func SetEventLogger(l log.Logger) EventOptionFunc {
	return func(c *EventClient) error {
		c.logger = l
		return nil
	}
}

func (c *EventClient) RecordEvent(event api.Event) (int64, error) {
	var (
		r      *sql.Rows
		result sql.Result
		err    error
		id     int64
	)
	if c.mysql {
		result, err = c.db.Exec(`INSERT INTO event
	(dt, username, passwd, remote_addr, remote_port, remote_name, remote_version, origin_addr, application, protocol)
        VALUES(?,?,?,?,?,?,?,?,?,?)
        `, event.Time, event.User, event.Passwd, event.RemoteAddr, event.RemotePort, event.RemoteName, event.RemoteVersion, event.OriginAddr, event.Application, event.Protocol)
		if err == nil {
			return result.LastInsertId()
		}
		return 0, err
	} else {
		r, err = c.db.Query(c.replaceParams(`INSERT INTO event
	(dt, username, passwd, remote_addr, remote_port, remote_name, remote_version, origin_addr, application, protocol)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING ID`),
			event.Time, event.User, event.Passwd, event.RemoteAddr, event.RemotePort, event.RemoteName, event.RemoteVersion, event.OriginAddr, event.Application, event.Protocol)
	}

	if err != nil {
		return 0, err
	}
	defer r.Close()
	r.Next()
	err = r.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (c *EventClient) GetEvent(id int64) (*api.EventGeo, error) {
	r := c.db.QueryRow(c.replaceParams(`SELECT
	id, dt, username, passwd, remote_addr, remote_name, remote_version, remote_port, remote_country, remote_city,
	origin_addr, origin_country, origin_city,
	remote_latitude, remote_longitude,
        origin_latitude, origin_longitude
	FROM event_geo WHERE id = ? LIMIT 1`), id)
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
	r, err := c.db.Query(c.replaceParams(`SELECT country_code,sum(hits) as hits FROM country_stats GROUP BY country_code ORDER BY country_code`))
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

var defaultLogger log.Logger

func (c *EventClient) replaceParams(sql string) string {
	if c.mysql {
		return sql
	}
	return replaceParams(sql)
}

func replaceParams(sql string) string {
	p := 1
	for strings.Contains(sql, "?") {
		sql = strings.Replace(sql, "?", fmt.Sprintf("$%d", p), 1)
		p++
	}
	return sql
}

func init() {
	defaultLogger = log.Logger{}
	defaultLogger.SetLevel(log.InfoLevel)
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
