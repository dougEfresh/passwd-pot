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
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/ftp"
	httppot "github.com/dougEfresh/passwd-pot/cmd/http"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"github.com/spf13/cobra"

	"fmt"
	"github.com/dougEfresh/passwd-pot/cmd/pop"
	"log/syslog"
	"net/http"
	"sync"
)

const (
	defaultFtpPort  = 2121
	defaultHttpPort = 8000
	defaultPopPort  = 1110
)

var potConfig struct {
	Ftp    int
	Http   int
	Telnet int
	Pop    int
	Vnc    int
	Health string
	Server string
	Bind   string
	All    bool
	DryRun bool
}

type potterClient struct {
	eventClient api.Transporter
}

type dryRunClient struct {
	eventClient api.Transporter
}

func (p *potterClient) Send(event *api.Event) {
	go func(e *api.Event) {
		log.Infof("Sending %s", e)
		if _, err := p.eventClient.SendEvent(e); err != nil {
			log.Errorf("Error sending event %s %s", e, err)
		}
	}(event)
}

func (d *dryRunClient) Send(event *api.Event) {

}
func (d *dryRunClient) SendEvent(event *api.Event) (*api.Event, error) {
	return nil, nil
}
func (d *dryRunClient) GetEvent(id int64) (*api.Event, error) {
	return nil, nil
}

// potterCmd represents the potter command
var potterCmd = &cobra.Command{
	Use:   "potter",
	Short: "potter",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if config.Debug {
			log.SetLevel(log.DebugLevel)
		}
		if config.Syslog != "" {
			if syslogHook, err := logrus_syslog.NewSyslogHook("tcp", config.Syslog, syslog.LOG_LOCAL0, "passwd-potter"); err != nil {
				log.Error("Unable to connect to local syslog daemon")
			} else {
				log.AddHook(syslogHook)
			}
		}
		if config.Pprof != "" {
			go func() { log.Error(http.ListenAndServe(config.Pprof, nil)) }()
		}
		runPotter()
	},
}

func runPotter() {
	log.Infof("Running pot client with %+v", potConfig)
	var pc *potterClient
	var err error
	var c api.Transporter
	if potConfig.DryRun {
		c = &dryRunClient{}
	} else {
		c, err = api.NewClient(potConfig.Server)
	}
	if err != nil {
		log.Panicf("Error creating eventCLient %s %s", potConfig.Server, err)
	}

	var wg sync.WaitGroup
	pc = &potterClient{
		eventClient: c,
	}
	if potConfig.All {
		wg.Add(1)
		go httppot.Run(&work.Worker{
			Addr:       fmt.Sprintf("%s:%d", potConfig.Bind, getPort(defaultHttpPort, potConfig.Http)),
			EventQueue: pc,
			Wg:         &wg,
		},
		)
		wg.Add(1)
		go ftp.Run(&work.Worker{
			Addr:       fmt.Sprintf("%s:%d", potConfig.Bind, getPort(defaultFtpPort, potConfig.Ftp)),
			EventQueue: pc,
			Wg:         &wg,
		},
		)
		wg.Add(1)
		go pop.Run(&work.Worker{
			Addr:       fmt.Sprintf("%s:%d", potConfig.Bind, getPort(defaultPopPort, potConfig.Pop)),
			EventQueue: pc,
			Wg:         &wg,
		},
		)
	}
	wg.Wait()
}

func getPort(defaultPort int, customPort int) int {
	if customPort > 0 {
		return customPort
	}
	return defaultPort
}

func init() {
	RootCmd.AddCommand(potterCmd)
	potterCmd.PersistentFlags().IntVar(&potConfig.Http, "http", 0, "create http pot")
	potterCmd.PersistentFlags().IntVar(&potConfig.Ftp, "ftp", 0, "create ftp pot")
	potterCmd.PersistentFlags().IntVar(&potConfig.Pop, "pop", 0, "create pop pot")
	potterCmd.PersistentFlags().IntVar(&potConfig.Vnc, "vnc", 0, "create vnc pot")
	potterCmd.PersistentFlags().IntVar(&potConfig.Telnet, "telnet", 0, "create ftp pot")
	potterCmd.PersistentFlags().StringVar(&potConfig.Server, "server", "http://localhost:8080", "send events to this server")
	potterCmd.PersistentFlags().StringVar(&potConfig.Bind, "bind", "localhost", "bind to this address")
	potterCmd.PersistentFlags().BoolVar(&potConfig.DryRun, "dry-run", false, "don't send events")
	potterCmd.PersistentFlags().BoolVar(&potConfig.All, "all", false, "run all potters")
}
