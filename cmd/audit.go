package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dougEfresh/dbr"
)

type AuditRecorder interface {
	RecordEvent(event *SshEvent) error
	ResolveGeoEvent(event *SshEvent) error
	Get(id int64) *SshEventGeo
}

type AuditClient struct {
	db *dbr.Connection
	geoClient GeoClientTransporter
}

func (ac *AuditClient) RecordEvent(event *SshEvent) error {
	log.Infof("Processing event %+v", event)
	var id int64
	sess := ac.db.NewSession(nil)
	err := sess.QueryRow(`INSERT INTO event(dt,username,passwd,remote_addr,remote_port,remote_name,remote_version,origin_addr)
	                            VALUES
	                           ($1,$2,$3,$4,$5,$6,$7,$8)
	                            RETURNING id`,
		event.getTime(), event.User, event.Passwd, event.RemoteAddr, event.RemotePort, event.RemoteName, event.RemoteVersion, event.OriginAddr).
		Scan(&id)
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

