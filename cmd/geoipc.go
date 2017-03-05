package cmd

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/gocraft/health"
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
	job := stream.NewJob("freegeo_lookup")
	res, err := http.Get(c.URL + "/" + ip)
	if err != nil {
		job.EventErr("freegeo_lookup_http_error", err)
		job.Complete(health.Error)
		return &Geo{}, err
	}

	var loc Geo
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&loc); err != nil {
		job.EventErr("freegeo_lookup_validation", err)
		job.Complete(health.ValidationError)
		return &Geo{}, err
	}
	loc.LastUpdate = time.Now()
	job.Complete(health.Success)
	return &loc, nil
}
