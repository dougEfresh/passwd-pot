package cmd

import (
	"encoding/json"
	"net/http"
	"time"
)

type GeoClientTransporter interface {
	GetLocationForIP(ip string) (*Geo, error)
}

type GeoClient struct {
	Url string
}

func DefaultGeoClient() *GeoClient {
	return &GeoClient{
		Url: "https://freegeoip.net/json",
	}
}

func (c *GeoClient) GetLocationForIP(ip string) (*Geo, error) {
	res, err := http.Get(c.Url + "/" + ip)
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
