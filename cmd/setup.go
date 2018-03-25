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
	"log/syslog"
	"net/http"
	"os"

	"github.com/dougEfresh/passwd-pot/log"
	"github.com/dougEfresh/zapz"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setup(cmd *cobra.Command, args []string) {
	if config.Pprof != "" {
		go func() { logger.Error(http.ListenAndServe(config.Pprof, nil)) }()
	}
	if len(args) > 0 {
		logger.Infof("Running %s with %s", cmd.Name(), args)
	} else {
		logger.Infof("Running - %s", cmd.Name())
	}
}

func setupLogger(name string) {
	logger = log.DefaultLogger(os.Stdout)
	logger.SetLevel(log.InfoLevel)
	h, _ := os.Hostname()
	if config.Debug {
		logger.SetLevel(log.DebugLevel)
	}
	if config.Syslog != "" {
		writer, err := syslog.Dial("tcp", config.Syslog, syslog.LOG_LOCAL0, name)
		if err != nil {
			en := zapcore.NewJSONEncoder(zapz.DefaultConfig)
			c := zapcore.NewCore(en, zapcore.AddSync(writer), zap.DebugLevel)
			logger.AddLogger(zap.New(c).With(zap.String("app", "default")))
		} else {
			//logger.AddLogger(klog.NewJSONLogger(writer))
		}
	}
	if config.Logz != "" {
		lz, err := zapz.New(config.Logz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to logz %s\n", err)
		} else {
			logger.AddLogger(lz)
		}
	}
	logger.With(zap.String("app", name))
	logger.With(zap.String("host", h))
}

var logger log.FieldLogger
