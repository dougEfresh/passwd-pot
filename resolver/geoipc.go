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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fiorix/freegeoip"
)

type GeoClientTransporter interface {
	GetLocationForAddr(ip string) (*Geo, error)
}

//GeoClient for geo IP
type GeoClient struct {
	URL string
}

type GeoClientDB struct {
	db *freegeoip.DB
}

func defaultGeoClient() *GeoClient {
	return &GeoClient{
		URL: "https://freegeoip.net/json",
	}
}

func (c *GeoClient) GetLocationForAddr(ip string) (*Geo, error) {
	res, err := http.Get(c.URL + "/" + ip)
	if err != nil {
		return &Geo{}, err
	}

	var loc Geo
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&loc); err != nil {
		return &Geo{}, err
	}
	loc.LastUpdate = time.Now()
	return &loc, nil
}

type geoipQuery struct {
	freegeoip.DefaultQuery
}

func (c *GeoClientDB) GetLocationForAddr(ip string) (*Geo, error) {
	var q geoipQuery
	ipaddr := net.ParseIP(ip)
	if ipaddr == nil {
		return nil, errors.New("Error parsing ip " + ip)
	}
	if err := c.db.Lookup(ipaddr, &q); err != nil {
		return nil, err
	}
	g := Geo{
		IP:          ip,
		LastUpdate:  time.Now(),
		CountryCode: q.Country.ISOCode,
		City:        q.City.Names["en"],
		TimeZone:    q.Location.TimeZone,
		Latitude:    q.Location.Latitude,
		Longitude:   q.Location.Longitude,
		MetroCode:   q.Location.MetroCode,
	}
	if len(q.Region) > 0 {
		g.RegionName = q.Region[0].Names["en"]
		g.RegionCode = q.Region[0].ISOCode
	}
	fmt.Printf("Returning %s", g)
	return &g, nil
}
