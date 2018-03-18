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

package resolver

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/dougEfresh/passwd-pot/cache"
	"github.com/dougEfresh/passwd-pot/potdb"

	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
	"github.com/fiorix/freegeoip"
)

// EventResolver interface
type EventResolver interface {
	ResolveEvent(event api.Event) ([]int64, error)
	MarkRemoteEvent(id int64, geoID int64) error
	MarkOriginEvent(id int64, geoID int64) error
}

type ResolveClient struct {
	db        potdb.DB
	geoClient GeoClientTransporter
	logger    log.FieldLogger
	useCache  bool
}

var geoCache = cache.NewCache()

type ResolveOptionFunc func(*ResolveClient) error

func NewResolveClient(options ...ResolveOptionFunc) (*ResolveClient, error) {
	rc := &ResolveClient{
		geoClient: defaultGeoClient(),
	}
	for _, option := range options {
		if err := option(rc); err != nil {
			return nil, err
		}
	}
	if rc.logger == nil {
		rc.logger = log.DefaultLogger(os.Stdout)
	}
	return rc, nil
}

func WithDsn(dsn string) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		var err error
		c.db, err = potdb.Open(dsn)
		return err
	}
}

func SetDb(db potdb.DB) ResolveOptionFunc {
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

func SetLogger(l log.FieldLogger) ResolveOptionFunc {
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

func UseCache() ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.useCache = true
		return nil
	}
}

func (c *ResolveClient) MarkOriginEvent(id int64, geoID int64) error {
	return c.markEvent(id, geoID, false)
}

func (c *ResolveClient) MarkRemoteEvent(id int64, geoID int64) error {
	return c.markEvent(id, geoID, true)
}

func (c *ResolveClient) markEvent(id int64, geoID int64, remote bool) error {
	var field = "remote"
	if !remote {
		field = "origin"
	}
	sql := fmt.Sprintf("UPDATE event SET %s_geo_id = ? where id = ?", field)
	c.logger.Debug(sql)
	_, err := c.db.Exec(sql, geoID, id)
	return err
}

func (c *ResolveClient) ResolveEvent(event api.Event) ([]int64, error) {
	var geoIds = []int64{0, 0}
	if c.useCache {
		rID, _ := geoCache.Get(event.RemoteAddr)
		oID, _ := geoCache.Get(event.OriginAddr)
		if rID > 0 && oID > 0 {
			var err error
			if e := c.MarkRemoteEvent(event.ID, rID); e != nil {
				err = e
			}
			if e := c.MarkOriginEvent(event.ID, oID); e != nil {
				err = e
			}
			return []int64{rID, oID}, err
		}
	}
	if event.ID == 0 {
		c.logger.Errorf("Got bad event: %s", event)
		return geoIds, errors.New("bad event recv")
	}
	var appendedErrors []error
	var err error
	var geoId int64
	if geoId, err = c.resolveAddr(event.RemoteAddr); err != nil {
		appendedErrors = append(appendedErrors, err)
	}
	if c.useCache && geoId > 0 {
		geoCache.Set(event.RemoteAddr, geoId)
	}
	geoIds[0] = geoId
	if err = c.MarkRemoteEvent(event.ID, geoId); err != nil {
		appendedErrors = append(appendedErrors, err)
	}
	if geoId, err = c.resolveAddr(event.OriginAddr); err != nil {
		appendedErrors = append(appendedErrors, err)
	}
	if c.useCache && geoId > 0 {
		geoCache.Set(event.RemoteAddr, geoId)
	}
	if err = c.MarkOriginEvent(event.ID, geoId); err != nil {
		appendedErrors = append(appendedErrors, err)
	}
	geoIds[1] = geoId
	if len(appendedErrors) > 0 {
		return geoIds, appendedErrors[0]
	}
	return geoIds, nil
}

func insertGeo(geo *Geo, db potdb.DB) (int64, error) {
	res, err := db.Insert(`INSERT INTO geo
	(ip, country_code, region_code, region_name, city, time_zone, latitude, longitude, metro_code, last_update)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		geo.IP, geo.CountryCode, geo.RegionCode, geo.RegionName, geo.City, geo.TimeZone, geo.Latitude, geo.Longitude, geo.MetroCode, geo.LastUpdate)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (c *ResolveClient) resolveAddr(addr string) (int64, error) {
	if m, _ := regexp.MatchString("\\d{1,3}\\.\\d{1,3}\\.", addr); !m {
		c.logger.Infof("%s is not an address", addr)
		addr = "127.0.0.1"
	}
	var geo = Geo{}
	var id int64
	expire := time.Now().AddDate(0, -1, 0)
	if c == nil {
		return 0, errors.New("FATAL: Resolve client is null")
	}
	r := c.db.QueryRow(`SELECT
	id, ip, country_code, region_code, region_name, city, time_zone, latitude, longitude, metro_code, last_update
	FROM geo
	WHERE ip = ?
	ORDER BY last_update DESC LIMIT 1`, addr)
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
		id, err = insertGeo(geo, c.db)
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
			if _, err = c.db.Exec("UPDATE geo SET last_update = now() WHERE id  = ?", geo.ID); err != nil {
				return 0, err
			}
		} else {
			c.logger.Infof("Inserting new record for id %d ", geo.ID)
			id, err = insertGeo(newGeo, c.db)
			if err != nil {
				return 0, err
			}
		}
	}
	return id, nil
}
