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
	klog "github.com/go-kit/kit/log"
	"github.com/gocraft/health"
	"github.com/newrelic/go-agent"
	"github.com/spf13/cobra"
	"io"
	"log/syslog"
	"net/http"
	"os"
)

var stream = health.NewStream()

func setup(cmd *cobra.Command, args []string) {
	var err error
	var writer io.Writer
	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}
	if config.Syslog != "" {
		if syslogHook, err = logrus_syslog.NewSyslogHook("tcp", config.Syslog, syslog.LOG_LOCAL0, cmd.Name()); err != nil {
			log.Error("Unable to connect to local syslog daemon")
		} else {
			log.AddHook(syslogHook)
		}
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
	if config.Pprof != "" {
		go func() { log.Error(http.ListenAndServe(config.Pprof, nil)) }()
	}
	if config.Syslog != "" {
		if syslogHook, err = logrus_syslog.NewSyslogHook("tcp", config.Syslog, syslog.LOG_LOCAL0, cmd.Name()); err != nil {
			log.Error("Unable to connect to local syslog daemon")
		} else {
			log.AddHook(syslogHook)
		}
		writer, err = syslog.Dial("tcp", config.Syslog, syslog.LOG_LOCAL0, cmd.Name())
	}

	if writer != nil {
		logger = klog.NewJSONLogger(writer)
	} else {
		logger = klog.NewJSONLogger(os.Stdout)
	}
	logger = klog.With(logger, "ts", klog.DefaultTimestampUTC)
	logger = klog.With(logger, "caller", klog.DefaultCaller)
}
