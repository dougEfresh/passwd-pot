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
	"github.com/fiorix/freegeoip"
	"regexp"
	"strings"
	"time"
)

type EventResolver interface {
	ResolveEvent(event api.Event) ([]int64, error)
	MarkEvent(id int64, geoId int64, remote bool) error
}

type ResolveClient struct {
	db        *sql.DB
	geoClient GeoClientTransporter
	logger    log.Logger
	mysql     bool
}

type ResolveOptionFunc func(*ResolveClient) error

func NewResolveClient(options ...ResolveOptionFunc) (*ResolveClient, error) {
	rc := &ResolveClient{
		logger:    defaultLogger,
		geoClient: defaultGeoClient(),
	}
	for _, option := range options {
		if err := option(rc); err != nil {
			return nil, err
		}
	}
	return rc, nil
}

func WithResolvDsn(dsn string) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		var err error
		c.db, err = loadDSN(dsn)
		c.mysql = !strings.Contains(dsn, "postgres")
		return err
	}
}

func SetResolveDb(db *sql.DB) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.db = db
		return nil
	}
}

func SetGeoDb(db string) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		geodb, err := freegeoip.Open(db)
		if err != nil {
			return err
		}
		c.geoClient = &GeoClientDB{
			db: geodb,
		}
		return nil
	}
}

func SetResolveLogger(l log.Logger) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.logger = l
		return nil
	}
}

func SetGeoClient(gc GeoClientTransporter) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.geoClient = gc
		return nil
	}
}

func (c *ResolveClient) MarkEvent(id int64, geoId int64, remote bool) error {

	if remote {
		c.logger.Debugf("Setting remote_id %d to %d ", geoId, id)
		_, err := c.db.Exec(c.replaceParams(`UPDATE event SET remote_geo_id = ? where id = ?`), geoId, id)
		return err
	}
	c.logger.Debugf("Setting origin_id %d to %d ", geoId, id)
	_, err := c.db.Exec(c.replaceParams(`UPDATE event SET origin_geo_id = ? where id = ?`), geoId, id)
	return err
}

func (c *ResolveClient) ResolveEvent(event api.Event) ([]int64, error) {
	var geoIds []int64 = []int64{0, 0}
	if event.ID == 0 {
		c.logger.Errorf("Got bad event: %s", event)
		return geoIds, errors.New("bad event recv")
	}
	var err error
	var geoId int64
	if geoId, err = c.resolveAddr(event.RemoteAddr); err != nil {
		return geoIds, err
	}
	if err = c.MarkEvent(event.ID, geoId, true); err != nil {
		return geoIds, err
	}
	geoIds[0] = geoId
	if geoId, err = c.resolveAddr(event.OriginAddr); err != nil {
		return geoIds, err
	}
	if err = c.MarkEvent(event.ID, geoId, false); err != nil {
		return geoIds, err
	}
	geoIds[1] = geoId
	return geoIds, nil
}

func insertGeo(geo *Geo, db *sql.DB, mysql bool) (int64, error) {
	var id int64
	if true {
		res, err := db.Exec(`INSERT INTO geo
	(ip, country_code, region_code, region_name, city, time_zone, latitude, longitude, metro_code, last_update)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			geo.IP, geo.CountryCode, geo.RegionCode, geo.RegionName, geo.City, geo.TimeZone, geo.Latitude, geo.Longitude, geo.MetroCode, geo.LastUpdate)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
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
	if m, _ := regexp.MatchString("\\d{1,3}\\.\\d{1,3}\\.", addr); !m {
		c.logger.Infof("%s is not an address", addr)
		addr = "127.0.0.1"
	}
	var geo = Geo{}
	var id int64 = 0
	expire := time.Now().AddDate(0, -1, 0)
	if c == nil {
		return 0, errors.New("FATAL: Resolve client is null")
	}
	r := c.db.QueryRow(c.replaceParams(`SELECT
	id, ip, country_code, region_code, region_name, city, time_zone, latitude, longitude, metro_code, last_update
	FROM geo
	WHERE ip = ?
	ORDER BY last_update DESC LIMIT 1`), addr)
	err := r.Scan(&geo.ID, &geo.IP, &geo.CountryCode, &geo.RegionCode, &geo.RegionName, &geo.City, &geo.TimeZone, &geo.Latitude, &geo.Longitude, &geo.MetroCode, &geo.LastUpdate)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	id = geo.ID
	if err == sql.ErrNoRows {
		c.logger.Infof("New addr found %s", addr)
		geo, err := c.geoClient.GetLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		id, err = insertGeo(geo, c.db, c.mysql)
		if err != nil {
			return 0, err
		}
	} else if geo.LastUpdate.Before(expire) {
		c.logger.Infof("Found expired addr %s (%s) (%s)", addr, geo.LastUpdate, expire)
		var newGeo = &Geo{}
		newGeo, err = c.geoClient.GetLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		if geo.equals(newGeo) {
			c.logger.Infof("Updating last_update for id %d ", geo.ID)
			if _, err = c.db.Exec(c.replaceParams("UPDATE geo SET last_update = now() WHERE id  = ?"), geo.ID); err != nil {
				return 0, err
			}
		} else {
			c.logger.Infof("Inserting new record for id %d ", geo.ID)
			id, err = insertGeo(newGeo, c.db, c.mysql)
			if err != nil {
				return 0, err
			}
		}
	}
	return id, nil
}

func (c *ResolveClient) replaceParams(sql string) string {
	return sql
	/*
		if c.mysql {
			return sql
		}
		return replaceParams(sql)
	*/
}
