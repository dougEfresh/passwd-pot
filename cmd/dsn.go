package cmd

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"strings"
)

const (
	eventTable    = "event"
	eventGeoTable = "event_geo"
	geoTable      = "geo"
)

func loadDSN(dsn string) *sql.DB {
	var db *sql.DB
	var err error
	if strings.Contains(dsn, "postgres") {
		log.Debug("Using pq driver")
		db, err = sql.Open("postgres", dsn)
	} else {
		log.Debug("Using mysql driver")
		db, err = sql.Open("mysql", dsn)
	}

	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
