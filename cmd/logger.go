package cmd

import (
	log "github.com/Sirupsen/logrus"
)

type kvs map[string]string

var dbEventLogger = &DbEvent{}

// DbEvent is a sentinel EventReceiver; use it if the caller doesn't supply one
type DbEvent struct{}

// Event receives a simple notification when various events occur
func (n *DbEvent) Event(eventName string) {
	log.Infof("eventName %s keyValueStore %s", eventName)
}

// EventKv receives a notification when various events occur along with
// optional key/value data
func (n *DbEvent) EventKv(eventName string, kvs map[string]string) {
	log.Infof("eventName %s keyValueStore %s", eventName, kvs)
}

// EventErr receives a notification of an error if one occurs
func (n *DbEvent) EventErr(eventName string, err error) error { return err }

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data
func (n *DbEvent) EventErrKv(eventName string, err error, kvs map[string]string) error {
	log.Errorf("eventName %s keyValueStore %s", eventName, kvs)
	return err
}

// Timing receives the time an event took to happen
func (n *DbEvent) Timing(eventName string, nanoseconds int64) {}

// TimingKv receives the time an event took to happen along with optional key/value data
func (n *DbEvent) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {}
