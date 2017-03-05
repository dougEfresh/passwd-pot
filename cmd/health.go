package cmd

import (
	"github.com/gocraft/health"
	"time"
	log "github.com/Sirupsen/logrus"
	"os"
)

var stream = health.NewStream()

func healthMonitor(name string) {
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

	if config.Statsd != "" {
		statsdOptions := &health.StatsDSinkOptions{
			Prefix: name,
		}
		statsdSink, err := health.NewStatsDSink(config.Statsd, statsdOptions)

		if err != nil {
			stream.EventErr("new_statsd_sink", err)
		} else {
			log.Infof("Configuring statsd at %s", config.Statsd)
			stream.AddSink(statsdSink)
		}
	}
}
