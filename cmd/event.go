package cmd

import (
	"encoding/binary"
	"time"
	"database/sql/driver"
)

type JsonTime time.Time

func (jt JsonTime) Unmarshal(data []byte, v interface{}) error {
	t := int64(binary.BigEndian.Uint64(data))
	time.Unix(t/1000, (t%1000)*1000000)
	v = &t
	return nil
}

// Value implements the driver Valuer interface.
func (jt JsonTime) Value() (driver.Value, error) {
	return time.Time(jt), nil
}

//SSHAudit data
type SshEventGeo struct {
	ID              int64     `db:"id"`
	Time            time.Time `db:"dt"`
	User            string    `db:"username"`
	Passwd          string    `db:"passwd"`
	RemoteAddr      string    `db:"remote_addr"`
	RemotePort      int       `db:"remote_port"`
	RemoteName      string    `db:"remote_name"`
	RemoteVersion   string    `db:"remote_version"`
	RemoteCountry   string    `db:"remote_country"`
	RemoteCity      string    `db:"remote_city"`
	OriginAddr      string    `db:"origin_addr"`
	OriginCountry   string    `db:"origin_country"`
	OriginCity      string    `db:"origin_city"`
	RemoteLatitude  float64   `db:"remote_latitude"`
	RemoteLongitude float64   `db:"remote_longitude"`
	OriginLatitude  float64   `db:"origin_latitude"`
	OriginLongitude float64   `db:"origin_longitude"`
	MetroCode       uint      `db:"metro_code"`
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
	ID            int64    `db:"id" json:"id"`
	Time          JsonTime `db:"dt" json:"time"`
	User          string   `db:"username"`
	Passwd        string   `db:"passwd"`
	RemoteAddr    string   `db:"remote_addr"`
	RemotePort    int      `db:"remote_port"`
	RemoteName    string   `db:"remote_name"`
	RemoteVersion string   `db:"remote_version"`
	OriginAddr    string   `db:"origin_addr"`
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
