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
	"testing"
)

func TestEventEquals(t *testing.T) {
	g := Geo{
		ID:          0,
		IP:          testEvent.RemoteAddr,
		CountryCode: "US",
		City:        "Mexico",
		RegionName:  "A galaxy far far away",
		RegionCode:  "666",
		TimeZone:    "GMT",
		Latitude:    1,
		Longitude:   2,
		MetroCode:   69,
	}

	another := Geo{
		ID:          0,
		IP:          testEvent.RemoteAddr,
		CountryCode: "US",
		City:        "New York",
		RegionName:  "A galaxy far far away",
		RegionCode:  "777",
		TimeZone:    "GMT",
		Latitude:    1,
		Longitude:   2,
		MetroCode:   69,
	}

	notEqual := Geo{
		ID:          0,
		IP:          testEvent.RemoteAddr,
		CountryCode: "CA",
		City:        "New York",
		RegionName:  "A galaxy far far away",
		RegionCode:  "777",
		TimeZone:    "GMT",
		Latitude:    1,
		Longitude:   2,
		MetroCode:   69,
	}

	if !g.equals(&g) {
		t.Fatalf("%+v != %+v", &g, &g)

	}

	if g.equals(&another) {
		t.Fatalf("%+v == %+v", &g, &another)

	}

	if g.equals(&notEqual) {
		t.Fatalf("%+v == %+v", &g, &notEqual)
	}
}
