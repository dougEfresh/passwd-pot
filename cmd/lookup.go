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
	log "github.com/Sirupsen/logrus"
	"gopkg.in/dougEfresh/dbr.v2"
	"sync"
	"time"
)

var eventChan = make(chan Event)
var geoCache *Cache

type mockGeoClient struct {
}

var mutex = &sync.Mutex{}
var geoClient = geoClientTransporter(defaultGeoClient())

func insertGeo(geo *Geo, session *dbr.Session) (int64, error) {
	var ids []int64
	_, err := session.InsertInto("geo").
		Columns("ip", "country_code", "region_code", "region_name", "city", "time_zone", "latitude", "longitude", "metro_code", "last_update").
		Record(geo).
		Returning(&ids, "id")
	if err != nil {
		return 0, err
	}
	geo.ID = ids[0]
	return geo.ID, nil
}

func (c *eventClient) resolveAddr(addr string) (int64, error) {
	if config.UseCache {
		cachedGeo, found := geoCache.Get(addr)
		if found {
			log.Debugf("Found cache hit for %s %d %d", addr, cachedGeo, geoCache.Count())
			return cachedGeo, nil
		}
	}

	mutex.Lock()
	defer mutex.Unlock()
	sess := c.db.NewSession(nil)
	expire := time.Now().AddDate(0, -1, 0)
	var geo = &Geo{}
	rowCount, err := sess.Select("*").
		From("geo").
		Where("ip = ?", addr).
		OrderDir("last_update", false).
		Limit(1).
		Load(&geo)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows || rowCount == 0 {
		log.Infof("New addr found %s", addr)
		geo, err = c.geoClient.getLocationForAddr(addr)
		if geo, err = c.geoClient.getLocationForAddr(addr); err != nil {
			log.Errorf("Error looking up IP: %s  %s", addr, err)
			return 0, err
		}
		if _, err = insertGeo(geo, sess); err != nil {
			return 0, err
		}
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
			if _, err = sess.UpdateBySql("UPDATE geo SET last_update = now() WHERE id = ?", geo.ID).
				Exec(); err != nil {
				return 0, err
			}
		} else {
			log.Infof("Inserting new record for id %d ", geo.ID)
			if _, err := insertGeo(newGeo, sess); err != nil {
				return 0, err
			}
			geo = newGeo

		}
	}
	geoCache.Set(addr, geo.ID)
	return geo.ID, nil
}

func runLookup() {
	log.Infof("Initalize lookup channel")
	for {
		select {
		case event := <-eventChan:
			go func(e Event) { defaultEventClient.resolveGeoEvent(&e) }(event)
		}
	}
}

func init() {
	geoCache = NewCache(5 * time.Minute)
}
