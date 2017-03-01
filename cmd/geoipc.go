package cmd

import (
	"encoding/json"
	"net/http"
	"time"
)

type geoClientTransporter interface {
	getLocationForAddr(ip string) (*Geo, error)
}

//GeoClient for geo IP
type GeoClient struct {
	URL string
}

func defaultGeoClient() *GeoClient {
	return &GeoClient{
		URL: "https://freegeoip.net/json",
	}
}

func (c *GeoClient) getLocationForAddr(ip string) (*Geo, error) {
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
