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
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

var eventChan = make(chan *Event)

type mockGeoClient struct {
}

var mutex = &sync.Mutex{}
var geoClient = geoClientTransporter(defaultGeoClient())
var geoPool = sync.Pool{
	New: func() interface{} {
		return &Geo{}
	},
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
	if !r.Next() {
		return 0, errors.New(fmt.Sprintf("Failed inserting %s", geo))
	}
	err = r.Scan(&id)
	return id, err
}

func (c *eventClient) resolveAddr(addr string) (int64, error) {
	geo := geoPool.Get().(*Geo)
	defer geoPool.Put(geo)
	mutex.Lock()
	defer mutex.Unlock()
	expire := time.Now().AddDate(0, -1, 0)
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
		log.Infof("New addr found %s", addr)
		geo, err = c.geoClient.getLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		return insertGeo(geo, c.db)
	}
	if geo.LastUpdate.Before(expire) {
		log.Infof("Found expired addr %s (%s) (%s)", addr, geo.LastUpdate, expire)
		var newGeo = &Geo{}
		newGeo, err = c.geoClient.getLocationForAddr(addr)
		if err != nil {
			return 0, err
		}
		if geo.equals(newGeo) {
			log.Infof("Updating last_update for id %d ", geo.ID)
			if _, err = c.db.Exec("UPDATE geo set last_update = now() where id  = $1", geo.ID); err != nil {
				return 0, err
			}
		} else {
			log.Infof("Inserting new record for id %d ", geo.ID)
			return insertGeo(newGeo, c.db)
			geo = newGeo
		}
	}
	return geo.ID, nil
}

func runLookup() {
	log.Infof("Initalize lookup channel")
	for {
		select {
		case event := <-eventChan:
			defaultEventClient.resolveGeoEvent(*event)
		}
	}
}

func init() {
	for i := 0; i < 100; i++ {
		geoPool.Put(&Geo{})
	}
}
