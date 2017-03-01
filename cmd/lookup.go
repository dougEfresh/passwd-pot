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
	"time"
)

type mockGeoClient struct {
}

var geoClient = GeoClientTransporter(DefaultGeoClient())

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

func (ac *AuditClient) resolveIp(ip string) (*Geo, error) {
	sess := ac.db.NewSession(nil)
	now := time.Now()
	expire := now.AddDate(0, -1, 0)
	var geo = &Geo{}

	rowCount, err := sess.Select("*").
		From("geo").
		Where("ip = ?", ip).
		OrderDir("last_update", false).
		Limit(1).
		Load(&geo)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows || rowCount == 0 {
		log.Infof("No rows returned for %s", ip)
		geo, err = ac.geoClient.GetLocationForIP(ip)
		if geo, err = ac.geoClient.GetLocationForIP(ip); err != nil {
			log.Errorf("Error looking up IP: %s  %s", ip, err)
			return nil, err
		}
		if _, err = insertGeo(geo, sess); err != nil {
			return nil, err
		}

		return geo, nil
	}
	if geo.LastUpdate.Before(expire) {
		var newGeo = &Geo{}
		newGeo, err = ac.geoClient.GetLocationForIP(ip)
		if err != nil {
			return nil, err
		}
		if geo.equals(newGeo) {
			log.Infof("Updating id %d ", geo.ID)
			if _, err = sess.UpdateBySql("UPDATE geo SET last_update = now() WHERE id = ?", geo.ID).
				Exec(); err != nil {
				return nil, err
			}
		} else {
			log.Infof("Inserting new record for id %d ", geo.ID)
			if _, err := insertGeo(newGeo, sess); err != nil {
				return nil, err
			}

		}
		log.Infof("Before %+v ", *geo)
	}
	log.Debugf("%+v", *geo)
	return geo, nil
}