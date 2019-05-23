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
	"net/http"
	"time"
)

type GeoClientTransporter interface {
	GetLocationForAddr(ip string) (*Geo, error)
}

//GeoClient for geo IP
type GeoClient struct {
	URL string
}

func (c *GeoClient) GetLocationForAddr(ip string) (*Geo, error) {
	res, err := http.Get(c.URL + "/json/" + ip)
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