package cmd

import (
	"github.com/gocraft/health"
	"time"
	log "github.com/Sirupsen/logrus"
	"os"
)

var stream = health.NewStream()

func healthMonitor() {
	if config.Debug {
		if syslogHook != nil {
			log.Infof("Configing syslog sinker")
			stream.AddSink(&health.WriterSink{syslogHook.Writer})
		} else {
			stream.AddSink(&health.WriterSink{os.Stdout})
		}
	}
	if config.Health != "" {
		log.Infof("Configuring health enpoint at %s", config.Health)
		jsonSink := health.NewJsonPollingSink(time.Minute, time.Minute*5)
		stream.AddSink(jsonSink)
		jsonSink.StartServer(config.Health)
	}
}
