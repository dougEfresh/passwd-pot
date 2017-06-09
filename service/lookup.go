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
	"errors"
	"fmt"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
	"time"
)

type EventResolver interface {
	ResolveEvent(event api.Event) ([]int64, error)
}

type ResolveClient struct {
	db        *sql.DB
	geoClient GeoClientTransporter
	log       log.Logger
}

type ResolveOptionFunc func(*ResolveClient) error

func NewResolveClient(options ...ResolveOptionFunc) (*ResolveClient, error) {
	rc := &ResolveClient{
		log: logger,
	}
	for _, option := range options {
		if err := option(rc); err != nil {
			return nil, err
		}
	}
	return rc, nil
}

func SetResolveDb(db *sql.DB) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.db = db
		return nil
	}
}

func SetResolveLogger(l log.Logger) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.log = l
		return nil
	}
}

func SetGeoClient(gc GeoClientTransporter) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.geoClient = gc
		return nil
	}
}

func (c *ResolveClient) ResolveEvent(event api.Event) ([]int64, error) {
	var geoIds []int64 = []int64{0, 0}
	if event.ID == 0 {
		err := errors.New("Bad event recv")
		logger.Errorf("Got bad event")
		return geoIds, err
	}
	var err error
	var geoId int64
	if geoId, err = c.resolveAddr(event.RemoteAddr); err != nil {
		return geoIds, err
	}
	if _, err = c.db.Exec(`UPDATE event SET remote_geo_id = $1 where id = $2`, geoId, event.ID); err != nil {
		return geoIds, err
	}
	geoIds[0] = geoId
	if geoId, err = c.resolveAddr(event.OriginAddr); err != nil {
		return geoIds, err
	}

	if _, err = c.db.Exec(`UPDATE event SET origin_geo_id = $1 where id = $2`, geoId, event.ID); err != nil {
		return geoIds, err
	}
	geoIds[1] = geoId
	return geoIds, nil
}

func insertGeo(geo *Geo, db *sql.DB) (int64, error) {
	var id int64
	r, err := db.Query(`INSERT INTO geo
	(ip, country_code, region_code, region_name, city, time_zone, latitude, longitude, metro_code, last_update)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING ID`,
		geo.IP, geo.CountryCode, geo.RegionCode, geo.RegionName, geo.City, geo.TimeZone, geo.Latitude, geo.Longitude, geo.MetroCode, geo.LastUpdate)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	if !r.Next() {
		return 0, errors.New(fmt.Sprintf("Failed inserting %s", geo))
	}
	err = r.Scan(&id)
	return id, err
}

func (c *ResolveClient) resolveAddr(addr string) (int64, error) {
	var geo = Geo{}
	var id int64 = 0
	expire := time.Now().AddDate(0, -1, 0)
	if c == nil {
		panic("no!")
	}
	r := c.db.QueryRow(`SELECT
	id, ip, country_code, region_code, region_name, city, time_zone, latitude, longitude, metro_code, last_update
	FROM geo
	WHERE ip = $1
	ORDER BY last_update DESC LIMIT 1`, addr)
	err := r.Scan(&geo.ID, &geo.IP, &geo.CountryCode, &geo.RegionCode, &geo.RegionName, &geo.City, &geo.TimeZone, &geo.Latitude, &geo.Longitude, &geo.MetroCode, &geo.LastUpdate)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		logger.Infof("New addr found %s", addr)
		geo, err := c.geoClient.GetLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		id, err = insertGeo(geo, c.db)
		if err != nil {
			return 0, err
		}
		logger.Infof("Adding %d to geo_event", id)
	} else if geo.LastUpdate.Before(expire) {
		logger.Infof("Found expired addr %s (%s) (%s)", addr, geo.LastUpdate, expire)
		var newGeo = &Geo{}
		newGeo, err = c.geoClient.GetLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		if geo.equals(newGeo) {
			logger.Infof("Updating last_update for id %d ", geo.ID)
			if _, err = c.db.Exec("UPDATE geo SET last_update = now() WHERE id  = $1", geo.ID); err != nil {
				return 0, err
			}
		} else {
			logger.Infof("Inserting new record for id %d ", geo.ID)
			id, err = insertGeo(newGeo, c.db)
			if err != nil {
				return 0, err
			}
		}
	}
	return id, nil
}
