package cmd

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/dougEfresh/dbr.v2"
)

type AuditRecorder interface {
	RecordEvent(event *SshEvent) error
	ResolveGeoEvent(event *SshEvent) error
	Get(id int64) *SshEventGeo
}

type AuditClient struct {
	db        *dbr.Connection
	geoClient GeoClientTransporter
}

func (ac *AuditClient) RecordEvent(event *SshEvent) error {
	log.Infof("Processing event %+v", event)
	sess := ac.db.NewSession(nil)
	var ids []int64
	_, err := sess.InsertInto("event").
		Columns("dt", "username", "passwd", "remote_addr", "remote_port", "remote_name", "remote_version", "origin_addr").
		Record(event).
		Returning(&ids, "id")

	if err != nil {
		return err
	}
	event.ID = ids[0]
	return nil
}

func (ac *AuditClient) ResolveGeoEvent(event *SshEvent) error {
	sess := ac.db.NewSession(nil)
	geo, err := ac.resolveIp(event.RemoteAddr)
	if err != nil {
		log.Errorf("Error geting location for RemoteAddr %+v %s", event, err)
		return err
	}
	updateBuilder := sess.Update("event").Set("remote_geo_id", geo.ID).Where("id = ?", event.ID)
	if _, err = updateBuilder.Exec(); err != nil {
		log.Errorf("Error updating remote_addr_geo_id for id %d %s", event.ID, err)
	}

	geo, err = ac.resolveIp(event.OriginAddr)
	if err != nil {
		log.Errorf("Errro getting location for origin %+v %s", event, err)
		return err
	}
	updateBuilder = sess.Update("event").Set("origin_geo_id", geo.ID).Where("id = ?", event.ID)
	if _, err = updateBuilder.Exec(); err != nil {
		log.Errorf("Error updating origin for id %d %s", event.ID, err)
		return err
	}
	return nil
}

func (ac *AuditClient) Get(id int64) *SshEventGeo {
	sess := ac.db.NewSession(nil)
	var event SshEventGeo
	if _, err := sess.Select("*").
		From("vw_event").
		Where("id = ?", id).
		Load(&event); err != nil {
		log.Errorf("Error getting event id %d %s", id, err)
		return nil
	}
	return &event
}
