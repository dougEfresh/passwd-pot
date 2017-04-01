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
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gocraft/health"
)

type eventRecorder interface {
	recordEvent(event Event) (int64, bool, error)
	resolveGeoEvent(event Event) error
}

type eventLister interface {
	list() []EventGeo
	get(id int64) *EventGeo
}

type eventTransporter interface {
	eventLister
	eventRecorder
}

type eventClient struct {
	db        *sql.DB
	geoClient geoClientTransporter
}

var defaultEventClient *eventClient

func (c *eventClient) list() []EventGeo {
	var geoEvents []EventGeo
	return geoEvents
}

func (c *eventClient) recordEvent(event Event) (int64, bool, error) {
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("Processing %s", event)
	}
	job := stream.NewJob("record_event")
	var r *sql.Rows
	var rId int64
	var oId int64
	var err error
	rId, _ = geoCache.Get(event.RemoteAddr)
	oId, _ = geoCache.Get(event.OriginAddr)
	if rId > 0 && oId > 0 {
		r, err = c.db.Query(`INSERT INTO event
	(dt, username, passwd, remote_addr, remote_geo_id, remote_port, remote_name, remote_version, origin_addr, origin_geo_id, application, protocol)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
        RETURNING ID`,
			event.Time, event.User, event.Passwd, event.RemoteAddr, rId, event.RemotePort, event.RemoteName, event.RemoteVersion, event.OriginAddr, oId, event.Application, event.Protocol)
	} else {
		r, err = c.db.Query(`INSERT INTO event
	(dt, username, passwd, remote_addr, remote_port, remote_name, remote_version, origin_addr, application, protocol)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING ID`,
			event.Time, event.User, event.Passwd, event.RemoteAddr, event.RemotePort, event.RemoteName, event.RemoteVersion, event.OriginAddr, event.Application, event.Protocol)
	}
	if err != nil {
		job.Complete(health.Error)
		return 0, false, err
	}
	var id int64
	defer r.Close()
	r.Next()
	err = r.Scan(&id)
	if err != nil {
		return 0, false, err
	}
	job.Complete(health.Success)
	return id, rId > 0 && oId > 0, nil
}

func (c *eventClient) resolveGeoEvent(event Event) error {
	job := stream.NewJob("resolve_geo_event")
	if event.ID == 0 {
		err := errors.New("Bad event recv")
		log.Errorf("Got bad event %s", event)
		job.EventErr("resolve_geo_event_invalid", err)
		job.Complete(health.ValidationError)
		return err
	}
	var err error
	var id int64
	if id, err = c.resolveAddr(event.RemoteAddr); err != nil {
		job.Complete(health.ValidationError)
		return err
	}
	if _, err = c.db.Exec(`UPDATE event SET remote_geo_id = $1 where id = $2`, id, event.ID); err != nil {
		job.Complete(health.Error)
		return err
	}

	if id, err = c.resolveAddr(event.OriginAddr); err != nil {
		job.Complete(health.Error)
		return err
	}

	if _, err = c.db.Exec(`UPDATE event SET origin_geo_id = $1 where id = $2`, id, event.ID); err != nil {
		job.Complete(health.Error)
		return err
	}
	job.Complete(health.Success)
	return nil
}

func (c *eventClient) broadcastEvent(id int64, hub *Hub) *EventGeo {
	if len(hub.clients) == 0 {
		return nil
	}
	gEvent := c.get(id)
	if gEvent == nil {
		return nil
	}
	if b, err := json.Marshal(gEvent); err != nil {
		log.Errorf("Error decoding geo event %d %s", id, err)
	} else {
		hub.broadcast <- b
	}
	return gEvent
}

func (c *eventClient) get(id int64) *EventGeo {
	job := stream.NewJob("get_event")
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
		log.Errorf("Error getting event id %d %s", id, err)
		job.Complete(health.Error)
		return nil
	}

	job.Complete(health.Success)
	return &event
}
