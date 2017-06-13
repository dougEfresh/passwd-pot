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
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/ftp"
	httppot "github.com/dougEfresh/passwd-pot/cmd/http"
	"github.com/dougEfresh/passwd-pot/cmd/pop"
	"github.com/dougEfresh/passwd-pot/cmd/psql"
	"github.com/dougEfresh/passwd-pot/cmd/queue"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"github.com/spf13/cobra"
	"net/http"
	"sync"
)

const (
	defaultFtpPort    = 2121
	defaultHttpPort   = 8080
	defaultPopPort    = 1110
	defaultTelnetPort = 2323
	defaultPsqlPort   = 5432
)

var potConfig struct {
	Ftp    int
	Http   int
	Telnet int
	Pop    int
	Psql   int
	Vnc    int
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
		logger.Infof("Sending %s", e)
		if err := p.eventClient.SendEvent(e); err != nil {
			logger.Errorf("Error sending event %s %s", e, err)
		}
	}(event)
}

func (d *dryRunClient) Send(event *api.Event) {

}
func (d *dryRunClient) SendEvent(event *api.Event) error {
	return nil
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
		setupLogger("passwd-potter")
		if config.Pprof != "" {
			go func() { logger.Error(http.ListenAndServe(config.Pprof, nil)) }()
		}
		runPotter()
	},
}

func runPotter() {
	logger.Infof("Running pot client with %+v", potConfig)
	var pc *potterClient
	var err error
	var c api.Transporter
	if potConfig.DryRun {
		c = &dryRunClient{}
	} else {
		c, err = api.NewClient(potConfig.Server)
	}
	if err != nil {
		logger.Errorf("Error creating eventCLient %s %s", potConfig.Server, err)
	}
	var wg sync.WaitGroup
	pc = &potterClient{
		eventClient: c,
	}

	if potConfig.All {
		wg.Add(1)
		go httppot.Run(getWorker(pc, &wg, getPort(defaultHttpPort, potConfig.Http), "http"), logger)
		wg.Add(1)
		go ftp.Run(getWorker(pc, &wg, getPort(defaultFtpPort, potConfig.Ftp), "ftp"), logger)
		wg.Add(1)
		go pop.Run(getWorker(pc, &wg, getPort(defaultPopPort, potConfig.Pop), "pop"), logger)
		wg.Add(1)
		go psql.Run(getWorker(pc, &wg, getPort(defaultPsqlPort, potConfig.Psql), "psql"), logger)
	} else {
		if potConfig.Http > 0 {
			wg.Add(1)
			go httppot.Run(getWorker(pc, &wg, getPort(defaultHttpPort, potConfig.Http), "http"), logger)
		}
		if potConfig.Ftp > 0 {
			wg.Add(1)
			go ftp.Run(getWorker(pc, &wg, getPort(defaultFtpPort, potConfig.Ftp), "ftp"), logger)
		}
	}

	wg.Wait()
}

func getWorker(eq queue.EventQueue, wg *sync.WaitGroup, port int, name string) work.Worker {

	return work.Worker{
		Addr:       fmt.Sprintf("%s:%d", potConfig.Bind, port),
		EventQueue: eq,
		Wg:         wg,
		Name:       name,
	}
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
	potterCmd.PersistentFlags().IntVar(&potConfig.Psql, "psql", 0, "create ftp pot")
	potterCmd.PersistentFlags().IntVar(&potConfig.Pop, "pop", 0, "create pop pot")
	potterCmd.PersistentFlags().IntVar(&potConfig.Vnc, "vnc", 0, "create vnc pot")
	potterCmd.PersistentFlags().IntVar(&potConfig.Telnet, "telnet", 0, "create ftp pot")
	potterCmd.PersistentFlags().StringVar(&potConfig.Server, "server", "http://localhost:8080", "send events to this server")
	potterCmd.PersistentFlags().StringVar(&potConfig.Bind, "bind", "localhost", "bind to this address")
	potterCmd.PersistentFlags().BoolVar(&potConfig.DryRun, "dry-run", false, "don't send events")
	potterCmd.PersistentFlags().BoolVar(&potConfig.All, "all", false, "run all potters")
}
