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
	log "github.com/Sirupsen/logrus"
	"github.com/dougEfresh/dbr"
	"time"
)

type mockGeoClient struct {
}

func (c *mockGeoClient) GetLocationForIP(ip string) (*Geo, error) {
	log.Infof("Looking up Geo Location for %s", ip)
	resp := []byte("{\"ip\":\"203.116.142.113\",\"country_code\":\"CA\",\"country_name\":\"Singapore\",\"region_code\":\"01\",\"region_name\":\"Central Singapore Community Development Council\",\"city\":\"Singapore\",\"zip_code\":\"\",\"time_zone\":\"Asia/Singapore\",\"latitude\":1.2855,\"longitude\":103.8565,\"metro_code\":0}")
	var geo = &Geo{}
	err := json.Unmarshal(resp, geo)
	geo.Ip = ip
	geo.LastUpdate = time.Now()
	return geo, err
}

//var geoClient = GeoClientTransporter(DefaultGeoClient())
var geoClient = GeoClientTransporter(&mockGeoClient{})

func insertGeo(geo *Geo, session *dbr.Session) (int64, error) {
	var id int64
	err := session.QueryRow(`
	INSERT INTO geo(ip, country_code, region_code, region_name, city, time_zone, latitude, longitude, metro_code, last_update)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	RETURNING id `,
		geo.Ip, geo.CountryCode, geo.RegionCode, geo.RegionName, geo.City, geo.TimeZone, geo.Latitude, geo.Longitude, geo.MetroCode, geo.LastUpdate).
		Scan(&id)

	if err != nil {
		return 0, err
	}
	geo.ID = id
	return id, nil
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

func (ac *AuditClient) ResolveGeoEvent(event *SshEvent) {
	sess := ac.db.NewSession(nil)
	geo, err := ac.resolveIp(event.RemoteAddr)
	if err != nil {
		log.Errorf("Error geting location for RemoteAddr %+v %s", event, err)
		return
	}
	updateBuilder := sess.Update("event").Set("remote_geo_id", geo.ID).Where("id = ?", event.ID)
	if _, err = updateBuilder.Exec(); err != nil {
		log.Errorf("Error updating remote_addr_geo_id for id %d %s", event.ID, err)
	}

	geo, err = ac.resolveIp(event.OriginAddr)
	if err !=nil {
		log.Errorf("Errro getting location for origin %+v %s", event, err)
		return
	}
	updateBuilder = sess.Update("event").Set("origin_geo_id", geo.ID).Where("id = ?", event.ID)
	if _, err = updateBuilder.Exec(); err != nil {
		log.Errorf("Error updating origin for id %d %s", event.ID, err)
	}
}
