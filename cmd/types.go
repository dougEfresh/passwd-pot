package cmd

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"errors"
	"strconv"
)

//Custom Serializer
type jsonTime struct {
	time.Time
}

// Time is in epoch ms
func (jt *jsonTime) UnmarshalJSON(data []byte) (err error) {
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return errors.New("could not decode time " + string(data))
	}
	jt.Time = time.Unix(ts/1000, (ts%1000)*1000000).UTC()
	return nil
}

func (jt jsonTime) MarshalJSON() ([]byte, error) {
	ft := jt.Time.UTC().UnixNano() / int64(time.Millisecond)
	return []byte(fmt.Sprintf("%d", ft)), nil
}

// Value implements the driver Valuer interface.
func (jt jsonTime) Value() (driver.Value, error) {
	return jt.Time, nil
}

// Gets the value from epoch time
func (n *jsonTime) Scan(value interface{}) error {
	var err error

	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		return nil
	case []byte:
		n.UnmarshalJSON(v)
		return nil
	case string:
		n.UnmarshalJSON([]byte(v))
		return err
	}

	return nil
}

//SSHEventGeo event with location
type SSHEventGeo struct {
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

//Geo location of addr
type Geo struct {
	ID          int64     `db:"id" json:"id"`
	IP          string    `db:"ip" json:"ip"`
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

//SSHEvent to record
type SSHEvent struct {
	ID            int64    `db:"id" json:"id"`
	Time          jsonTime `db:"dt" json:"time"`
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
		g.TimeZone == another.TimeZone &&
		g.IP == another.IP
}

func (e SSHEvent) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return string(b)
}
