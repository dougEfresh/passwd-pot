package cmd

import (
	"github.com/dougEfresh/dbr"
	"time"
)

//SSHAudit data
type SshEventGeo struct {
	ID            int64     `db:"id"`
	Time          time.Time `db:"dt"`
	User          string    `db:"username"`
	Passwd        string    `db:"passwd"`
	RemoteAddr    string    `db:"remote_addr"`
	RemotePort    int       `db:"remote_port"`
	RemoteName    string    `db:"remote_name"`
	RemoteVersion string    `db:"remote_version"`
	Longitude     float64   `db:"remote_longitude"`
	Latitude      float64   `db:"remote_latitude"`
	Country       string    `db:"country_code"`
	City          string    `db:"city"`
	MetroCode     uint      `db:"metro_code"`
}

type Geo struct {
	ID          int64     `db:"id" json:"id"`
	Ip          string    `db:"ip" json:"ip"`
	LastUpdate  time.Time `db:"last_update" json:"last_update"`
	CountryCode string    `db:"country_code" json:"country_code"`
	RegionCode  string    `db:"region_code" json:"region_code"`
	RegionName  string    `db:"region_name" json:"region_name"`
	City        string    `db:"city" json:"city"`
	TimeZone    string    `db:"time_zone" json:"time_zone"`
	Latitude    float64   `db:"latitude" json:"latitude"`
	Longitude   float64   `db:"longitude" json:"longitude"`
	MetroCode   int       `db:"metro_code" json:"metro_code"`
}

type SshEvent struct {
	ID              int64
	Epoch           int64  `db:"-" json:"time"`
	User            string `db:"username"`
	Passwd          string
	RemoteAddr      string
	RemoteAddrGeoId dbr.NullInt64 `db:"remote_geo_id" json:"-"`
	RemotePort      int
	RemoteName      string
	RemoteVersion   string
	OriginAddr      string
	OriginGeoId     dbr.NullInt64 `db:"origin_geo_id" json:"-"`
}

func (e *SshEvent) getTime() time.Time {
	return time.Unix(e.Epoch/1000, (e.Epoch%1000)*1000000)
}

func (g *Geo) equals(another *Geo) bool {
	return g.CountryCode == another.CountryCode &&
		g.City == another.City &&
		g.Latitude == another.Latitude &&
		g.Longitude == another.Longitude &&
		g.MetroCode == another.MetroCode &&
		g.RegionName == another.RegionName &&
		g.RegionCode == another.RegionCode &&
		g.TimeZone == another.TimeZone
}
