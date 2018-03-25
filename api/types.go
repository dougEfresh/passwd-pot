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

package api

import (
	"time"
)

// EventTime Custom Serializer
type EventTime time.Time

//Event to record
type Event struct {
	ID            int64
	Time          EventTime `json:"time"`
	User          string    `json:"user"`
	Passwd        string    `json:"passwd"`
	RemoteAddr    string    `json:"remoteAddr"`
	RemoteGeoID   int64     `json:"remoteGeoId"`
	RemotePort    int       `json:"remotePort"`
	RemoteName    string    `json:"remoteName"`
	RemoteVersion string    `json:"remoteVersion"`
	OriginAddr    string    `json:"originAddr"`
	OriginGeoID   int64     `json:"originGeoId"`
	Application   string    `json:"application"`
	Protocol      string    `json:"protocol"`
}

//CountryStat stats on cc
type CountryStat struct {
	Country   string
	Latitude  float64
	Longitude float64
	Hits      int64
}

//EventGeo event with location
type EventGeo struct {
	ID              int64
	Time            time.Time
	User            string
	Passwd          string
	RemoteAddr      string
	RemotePort      int
	RemoteName      string
	RemoteVersion   string
	RemoteCountry   string
	RemoteCity      string
	OriginAddr      string
	OriginCountry   string
	OriginCity      string
	RemoteLatitude  float64
	RemoteLongitude float64
	OriginLatitude  float64
	OriginLongitude float64
	MetroCode       uint
}

// BatchEventResponse response for batch inserts
type BatchEventResponse struct {
	Duration time.Duration `json:"duration"`
	Rows     int64         `json:"rows"`
	LastID   int64         `json:"lastId"`
}

// Transporter for event Client
type Transporter interface {
	ReadTransporter
	RecordTransporter
}

// ReadTransporter for data
type ReadTransporter interface {
	GetEvent(id int64) (*EventGeo, error)
}

// RecordTransporter Methods to records password events
type RecordTransporter interface {
	RecordEvent(event Event) (int64, error)
	RecordBatchEvents(events []Event) (BatchEventResponse, error)
}
