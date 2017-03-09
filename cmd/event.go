package cmd

import (
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gocraft/health"
	"gopkg.in/dougEfresh/dbr.v2"
	"time"
)

type eventRecorder interface {
	recordEvent(event *Event) error
	resolveGeoEvent(event *Event) error
}

type eventLister interface {
	list() []EventGeo
	get(id int64) *EventGeo
}

type eventTransporter interface {
	eventLister
	eventRecorder
}

type eventClient struct {
	db        *dbr.Connection
	geoClient geoClientTransporter
}

var defaultEventClient *eventClient

func (c *eventClient) list() []EventGeo {
	sess := c.db.NewSession(nil)
	after := time.Now().UTC().AddDate(0, 0, -1)
	var geoEvents []EventGeo
	_, err := sess.Select("*").
		From("vw_event").
		Where("dt > ?", after).
		Limit(1000).
		LoadValues(&geoEvents)
	if err != nil {
		log.Errorf("Error getting events %s", err)
	}
	return geoEvents
}

func (c *eventClient) recordEvent(event *Event) error {
	log.Infof("Processing event %+v", event)
	job := stream.NewJob("record_event")
	if c.db == nil {
		return nil
	}
	sess := c.db.NewSession(nil)
	var ids []int64
	_, err := sess.InsertInto("event").
		Columns("dt", "username", "passwd", "remote_addr", "remote_port", "remote_name", "remote_version", "origin_addr", "application", "protocol").
		Record(event).
		Returning(&ids, "id")

	if err != nil {
		job.Complete(health.Error)
		return err
	}
	event.ID = ids[0]
	job.Complete(health.Success)
	return nil
}

func (c *eventClient) resolveGeoEvent(event *Event) error {
	job := stream.NewJob("resolve_geo_event")
	if event.ID == 0 {
		err := errors.New("Bad event recv")
		log.Errorf("Got bad event %s", event)
		job.EventErr("resolve_geo_event_invalid", err)
		job.Complete(health.ValidationError)
		return err
	}

	sess := c.db.NewSession(nil)
	geo, err := c.resolveAddr(event.RemoteAddr)
	if err != nil {
		log.Errorf("Error geting location for RemoteAddr %+v %s", event, err)
		job.Complete(health.ValidationError)
		return err
	}
	updateBuilder := sess.Update("event").Set("remote_geo_id", geo.ID).Where("id = ?", event.ID)
	if _, err = updateBuilder.Exec(); err != nil {
		log.Errorf("Error updating remote_addr_geo_id for id %d %s", event.ID, err)
		job.Complete(health.Error)
		return err
	}

	geo, err = c.resolveAddr(event.OriginAddr)
	if err != nil {
		log.Errorf("Errro getting location for origin %+v %s", event, err)
		job.Complete(health.Error)
		return err
	}
	updateBuilder = sess.Update("event").Set("origin_geo_id", geo.ID).Where("id = ?", event.ID)
	if _, err = updateBuilder.Exec(); err != nil {
		log.Errorf("Error updating origin for id %d %s", event.ID, err)
		job.Complete(health.Error)
		return err
	}
	job.Complete(health.Success)
	go c.broadcastEvent(event.ID)
	return nil
}

func (c *eventClient) broadcastEvent(id int64) {
	gEvent := c.get(id)
	if gEvent == nil {
		return
	}
	if b, err := json.Marshal(gEvent); err != nil {
		log.Errorf("Error decoding geo event %d %s", id, err)
	} else {
		hub.broadcast <- b
	}
}

func (c *eventClient) get(id int64) *EventGeo {
	job := stream.NewJob("get_event")
	sess := c.db.NewSession(nil)
	var event EventGeo
	if _, err := sess.Select("*").
		From("vw_event").
		Where("id = ?", id).
		Load(&event); err != nil {
		log.Errorf("Error getting event id %d %s", id, err)
		job.Complete(health.Error)
		return nil
	}
	job.Complete(health.Success)
	return &event
}
