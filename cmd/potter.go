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
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/ftp"
	"github.com/dougEfresh/passwd-pot/cmd/http"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"github.com/spf13/cobra"
	"sync"
)

var potConfig struct {
	Ftp    string
	Http   string
	Telnet string
	Vnc    string
	Health string
	Server string
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
	Run:   func(cmd *cobra.Command, args []string) { runPotter() },
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
	wg.Add(1)
	go http.Run(&work.Worker{
		Addr:       potConfig.Http,
		EventQueue: pc,
		Wg:         &wg,
	},
	)
	wg.Add(1)
	go ftp.Run(&work.Worker{
		Addr:       potConfig.Ftp,
		EventQueue: pc,
		Wg:         &wg,
	},
	)
	wg.Wait()
}

func init() {
	RootCmd.AddCommand(potterCmd)
	potterCmd.PersistentFlags().StringVar(&potConfig.Http, "http", "", "create http pot")
	potterCmd.PersistentFlags().StringVar(&potConfig.Ftp, "ftp", "", "create ftp pot")
	potterCmd.PersistentFlags().StringVar(&potConfig.Vnc, "vnc", "", "create vnc pot")
	potterCmd.PersistentFlags().StringVar(&potConfig.Telnet, "telnet", "", "create ftp pot")
	potterCmd.PersistentFlags().StringVar(&potConfig.Server, "server", "http://localhost:8080", "send events to this server")
	potterCmd.PersistentFlags().BoolVar(&potConfig.DryRun, "dry-run", false, "don't send events")
}
