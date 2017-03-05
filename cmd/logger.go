package cmd

import (
	"github.com/Sirupsen/logrus/hooks/syslog"
)

var syslogHook *logrus_syslog.SyslogHook

var defaultDbEventLogger = &DbEvent{}

// DbEvent is a sentinel EventReceiver; use it if the caller doesn't supply one
type DbEvent struct{
	Debug bool
}

// Event receives a simple notification when various events occur
func (n *DbEvent) Event(eventName string) {
	if n.Debug {
		stream.Event(eventName)
	}
}

// EventKv receives a notification when various events occur along with
// optional key/value data
func (n *DbEvent) EventKv(eventName string, kvs map[string]string) {
	if n.Debug {
		stream.Event(eventName)
	}
}

// EventErr receives a notification of an error if one occurs
func (n *DbEvent) EventErr(eventName string, err error) error {
	stream.EventErr(eventName,err)
	return err
}

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data
func (n *DbEvent) EventErrKv(eventName string, err error, kvs map[string]string) error {
	stream.EventErrKv(eventName, err, kvs)
	return err
}

// Timing receives the time an event took to happen
func (n *DbEvent) Timing(eventName string, nanoseconds int64) {
	stream.Timing(eventName, nanoseconds)
}

// TimingKv receives the time an event took to happen along with optional key/value data
func (n *DbEvent) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	stream.TimingKv(eventName, nanoseconds, kvs)
}
