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
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
)

type EventRecorder interface {
	RecordEvent(event api.Event) (int64, error)
}

type EventLister interface {
	Get(id int64) *EventGeo
}

type EventTransporter interface {
	EventLister
	EventRecorder
}

type eventClient struct {
	db *sql.DB
}

var defaultEventClient *eventClient

func (c *eventClient) RecordEvent(event api.Event) (int64, error) {
	var (
		r   *sql.Rows
		err error
		id  int64
	)
	r, err = c.db.Query(`INSERT INTO event
	(dt, username, passwd, remote_addr, remote_port, remote_name, remote_version, origin_addr, application, protocol)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING ID`,
		event.Time, event.User, event.Passwd, event.RemoteAddr, event.RemotePort, event.RemoteName, event.RemoteVersion, event.OriginAddr, event.Application, event.Protocol)

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

func (c *eventClient) Get(id int64) *EventGeo {
	r := c.db.QueryRow(`SELECT
	id, dt, username, passwd, remote_addr, remote_name, remote_version, remote_port, remote_country, remote_city,
	origin_addr, origin_country, origin_city,
	remote_latitude, remote_longitude,
        origin_latitude, origin_longitude
	FROM event_geo WHERE id = $1 LIMIT 1`, id)
	var event EventGeo
	err := r.Scan(&event.ID, &event.Time, &event.User, &event.Passwd,
		&event.RemoteAddr, &event.RemoteName, &event.RemoteVersion,
		&event.RemotePort, &event.RemoteCountry, &event.RemoteCity,
		&event.OriginAddr, &event.OriginCountry, &event.OriginCity,
		&event.RemoteLatitude, &event.RemoteLongitude,
		&event.OriginLatitude, &event.OriginLongitude)
	if err != nil {
		logger.Errorf("Error getting event id %d %s", id, err)
		return nil
	}

	return &event
}

var logger log.Logger

func init() {
	logger = log.Logger{}
	logger.SetLevel(log.InfoLevel)
}
