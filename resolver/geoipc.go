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
	"time"

	"github.com/qioalice/ipstack"
)

type GeoClientTransporter interface {
	GetLocationForAddr(ip string) (*Geo, error)
}

//GeoClient for geo IP
type GeoClient struct {
	Token string
	URL string
}


func DefaultGeoClient(token string) (*GeoClient,  error) {
	err := ipstack.Init(ipstack.ParamToken(token), ipstack.ParamDisableFirstMeCall(), ipstack.ParamUseHTTPS(false))
	if err != nil {
		return nil, err
	}
	return &GeoClient{
		 Token: token,
		 URL: "http://api.ipstack.com",
	}, nil
}

func (c *GeoClient) GetLocationForAddr(ip string) (*Geo, error) {

	res, err := ipstack.IP(ip)
	if err != nil {
		return &Geo{}, err
	}
	var loc Geo
	loc.City = res.City
	loc.CountryCode = res.CountryCode
	loc.IP = ip
	loc.LastUpdate = time.Now()
	loc.Latitude = float64(res.Latitide)
	loc.Longitude = float64(res.Longitude)
	loc.MetroCode = 0
	loc.RegionName = res.RegionName
	loc.RegionCode = res.RegionCode
	if res.Timezone != nil {
		loc.TimeZone = res.Timezone.Code
	}
	return &loc, nil
}