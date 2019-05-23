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
	"regexp"
	"time"

	"github.com/dougEfresh/passwd-pot/cache"
	"github.com/dougEfresh/passwd-pot/potdb"
	"go.uber.org/multierr"

	"github.com/dougEfresh/passwd-pot/api"
)

// EventResolver interface
type EventResolver interface {
	ResolveEvent(event api.Event) ([]int64, error)
	Resolve(addr string) (int64, error)
	MarkRemoteEvent(id int64, geoID int64) error
	MarkOriginEvent(id int64, geoID int64) error
}

type ResolveClient struct {
	db        potdb.DB
	geoClient GeoClientTransporter
	useCache  bool
}

var geoCache = cache.NewCache()

type ResolveOptionFunc func(*ResolveClient) error

func NewResolveClient(options ...ResolveOptionFunc) (*ResolveClient, error) {
	rc := &ResolveClient{

	}
	for _, option := range options {
		if err := option(rc); err != nil {
			return nil, err
		}
	}
	if rc.geoClient == nil {
		return nil, errors.New("no geo client available")
	}
	return rc, nil
}

func SetDb(db potdb.DB) ResolveOptionFunc {
	return func(c *ResolveClient) error {
		c.db = db
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

func (c *ResolveClient) Resolve(addr string) (int64, error) {
	if c.useCache {
		id, ok := geoCache.Get(addr)
		if ok {
			return id, nil
		}
	}
	return c.resolveAddr(addr)
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
	_, err := c.db.Exec(sql, geoID, id)
	return err
}

func (c *ResolveClient) ResolveEvent(event api.Event) ([]int64, error) {
	var geoIds = []int64{0, 0}
	if event.ID == 0 {
		return geoIds, errors.New("bad event recv")
	}
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

	var rerr error
	var err error
	var geoId int64
	if geoId, err = c.resolveAddr(event.RemoteAddr); err != nil {
		rerr = multierr.Append(rerr, err)
	}
	if c.useCache && geoId > 0 {
		geoCache.Set(event.RemoteAddr, geoId)
	}
	geoIds[0] = geoId
	if err = c.MarkRemoteEvent(event.ID, geoId); err != nil {
		rerr = multierr.Append(rerr, err)
	}
	if geoId, err = c.resolveAddr(event.OriginAddr); err != nil {
		rerr = multierr.Append(rerr, err)
	}
	if c.useCache && geoId > 0 {
		geoCache.Set(event.RemoteAddr, geoId)
	}
	if err = c.MarkOriginEvent(event.ID, geoId); err != nil {
		rerr = multierr.Append(rerr, err)
	}
	geoIds[1] = geoId
	return geoIds, rerr
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
		geo, err := c.geoClient.GetLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		id, err = insertGeo(geo, c.db)
		if err != nil {
			return 0, err
		}
	} else if geo.LastUpdate.Before(expire) {
		var newGeo = &Geo{}
		newGeo, err = c.geoClient.GetLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		if geo.equals(newGeo) {
			if _, err = c.db.Exec("UPDATE geo SET last_update = now() WHERE id  = ?", geo.ID); err != nil {
				return 0, err
			}
		} else {
			id, err = insertGeo(newGeo, c.db)
			if err != nil {
				return 0, err
			}
		}
	}
	return id, nil
}
