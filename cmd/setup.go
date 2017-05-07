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
	"fmt"
	"github.com/dougEfresh/kitz"
	"github.com/dougEfresh/passwd-pot/cmd/log"
	klog "github.com/go-kit/kit/log"
	"github.com/newrelic/go-agent"
	"github.com/spf13/cobra"
	"log/syslog"
	"net/http"
	"os"
)

func setup(cmd *cobra.Command, args []string) {
	var err error
	if config.NewRelic != "" {
		config := newrelic.NewConfig("passwd-pot", config.NewRelic)
		if app, err = newrelic.NewApplication(config); err != nil {
			logger.Errorf("Could not start new relic %s", err)
		}
		logger.Infof("Configured new relic agent")
	}
	if config.Pprof != "" {
		go func() { logger.Error(http.ListenAndServe(config.Pprof, nil)) }()
	}
	defaultEventClient = &eventClient{
		db:        loadDSN(config.Dsn),
		geoClient: geoClient,
	}
	if len(args) > 0 {
		logger.Infof("Running %s with %s", cmd.Name(), args)
	} else {
		logger.Infof("Running %s", cmd.Name())
	}
}

func setupLogger(name string) {
	h, _ := os.Hostname()
	if config.Debug {
		logger.SetLevel(log.DebugLevel)
	}
	if config.Syslog != "" {
		writer, err := syslog.Dial("tcp", config.Syslog, syslog.LOG_LOCAL0, name)
		if err != nil {
			logger.AddLogger(klog.NewJSONLogger(os.Stdout))
			logger.Errorf("syslog failed %s", err)
		} else {
			logger.AddLogger(klog.NewJSONLogger(writer))
		}
	} else {
		logger.AddLogger(klog.NewJSONLogger(os.Stdout))
	}
	if config.Logz != "" {
		lz, err := kitz.New(config.Logz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to logz %s\n", err)
		} else {
			logger.AddLogger(lz)
		}
	}

	logger.With("app", name)
	logger.With("host", h)
	logger.With("ts", klog.DefaultTimestampUTC)
	logger.With("caller", klog.Caller(4))
}

var logger log.Logger

func init() {
	logger = log.Logger{}
	logger.SetLevel(log.InfoLevel)
}
