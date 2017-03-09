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
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	//DB driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type task interface {
	execute()
}

func (e Event) execute() {
	if e.ID%1000 == 0 {
		log.Infof("Running %s", e)
	}

	b, err := json.Marshal(e)
	if err != nil {
		log.Errorf("Error decoding %s", err)
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s%s", config.BindAddr, eventURL),
		"application/json", bytes.NewReader(b))

	if err != nil {
		log.Errorf("Error posting %s", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		log.Errorf("Error posting %s", resp.Status)
	}
	time.Sleep(500 * time.Millisecond)
}

type pool struct {
	mu    sync.Mutex
	size  int
	tasks chan task
	kill  chan struct{}
	wg    sync.WaitGroup
}

func newPool(size int) *pool {
	pool := &pool{
		tasks: make(chan task, 128),
		kill:  make(chan struct{}),
	}
	pool.resize(size)
	return pool
}

func (p *pool) worker() {
	defer p.wg.Done()
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			task.execute()
		case <-p.kill:
			return
		}
	}
}

func (p *pool) resize(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.size < n {
		p.size++
		p.wg.Add(1)
		go p.worker()
	}
	for p.size > n {
		p.size--
		p.kill <- struct{}{}
	}
}

func (p *pool) Close() {
	close(p.tasks)
}

func (p *pool) Wait() {
	p.wg.Wait()
}

func (p *pool) Exec(task task) {
	p.tasks <- task
}

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if config.Debug {
			log.SetLevel(log.DebugLevel)
		}
		db := loadDSN(config.Dsn)
		runtime.GOMAXPROCS(5)
		sess := db.NewSession(nil)
		var events []Event
		log.Info("Running query")
		num, err := sess.Select("*").
			From("event").
			//			Where("id > ?", 99624).
			OrderBy("id").LoadValues(&events)
		if err != nil {
			log.Errorf("Error running query %s ", err)
		}
		log.Info("Done running query (%d)", num)
		p := newPool(10)
		for _, e := range events {
			p.Exec(e)
		}
		log.Info("Closing channel")
		p.Close()
		log.Info("Waiting for workings")
		p.Wait()
	},
}

func init() {
	RootCmd.AddCommand(queryCmd)
}
