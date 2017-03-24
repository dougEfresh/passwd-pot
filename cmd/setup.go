// Copyright © 2017 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/gocraft/health"
	"github.com/newrelic/go-agent"
	"github.com/spf13/cobra"
	"log/syslog"
	"net/http"
	"os"
	"time"
)

var stream = health.NewStream()

func setup(cmd *cobra.Command, args []string) {
	var err error
	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}
	defaultEventClient = &eventClient{
		db:        loadDSN(config.Dsn),
		geoClient: geoClient,
	}
	log.Debugf("Running %s with %s", cmd.Name(), args)
	if config.NewRelic != "" {
		config := newrelic.NewConfig("passwd-pot", config.NewRelic)
		if app, err = newrelic.NewApplication(config); err != nil {
			log.Errorf("Could not start new relic %s", err)
		}
		log.Infof("Configured new relic agent")
	}
	if config.Syslog != "" {
		if syslogHook, err = logrus_syslog.NewSyslogHook("tcp", config.Syslog, syslog.LOG_LOCAL0, "passwd-pot"); err != nil {
			log.Error("Unable to connect to local syslog daemon")
		} else {
			log.AddHook(syslogHook)
		}
	}
	if config.Pprof != "" {
		go func() { log.Error(http.ListenAndServe(config.Pprof, nil)) }()
	}

	defaultDbEventLogger.Debug = config.Debug
	healthMonitor(cmd.Name())
}

func healthMonitor(name string) {
	if config.Debug {
		if syslogHook != nil {
			log.Infof("Configuring syslog sinker")
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
