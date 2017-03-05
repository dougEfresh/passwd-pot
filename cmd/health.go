package cmd

import (
	"github.com/gocraft/health"
	"time"
	log "github.com/Sirupsen/logrus"
)

var stream = health.NewStream()

func healthMonitor() {
	if hook != nil {
		stream.AddSink(&health.WriterSink{hook.Writer})
	}
	if config.Health == "" {
		return
	}
	log.Infof("Configuring health enpoint at %s", config.Health)
	jsonSink := health.NewJsonPollingSink(time.Minute, time.Minute*5)
	stream.AddSink(jsonSink)
	jsonSink.StartServer(config.Health)
}
