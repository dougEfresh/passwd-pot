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
	"encoding/json"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/beeker1121/goque"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/spf13/cobra"

	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	maxSize = 9 * 1024 * 1024
)

type socketRelay struct {
	q             *goque.Queue
	drainDuration time.Duration
	mux           sync.Mutex
	c             api.RecordTransporter
}

type socketDryRun struct {
	events []api.Event
}

func (d *socketDryRun) RecordEvent(event api.Event) (int64, error) {
	return 0, nil
}

func (d *socketDryRun) RecordBatchEvents(events []api.Event) (api.BatchEventResponse, error) {
	d.events = events
	logger.Infof("Sending %d events", len(d.events))
	return api.BatchEventResponse{}, nil
}

var sockerDryRunner = &socketDryRun{}
var socketRelayer = &socketRelay{}

var socketConfig struct {
	Pprof    string
	Server   string
	Socket   string
	DryRun   bool
	Duration time.Duration
}

func cobrearun(cmd *cobra.Command, args []string) {
	run(cmd.Name())
}

func run(name string) {
	setupLogger(name)
	if config.Pprof != "" {
		go func() { logger.Error(http.ListenAndServe(config.Pprof, nil)) }()
	}
	q, err := goque.OpenQueue(fmt.Sprintf("%s%s%s%s%d", os.TempDir(), string(os.PathSeparator), name, string(os.PathSeparator), time.Now().UnixNano()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error setting up queue %s", err)
		os.Exit(1)
	}

	socketRelayer.q = q
	socketRelayer.drainDuration = socketConfig.Duration
	if socketConfig.DryRun {
		socketRelayer.c = sockerDryRunner
	} else {
		c, err := api.New(socketConfig.Server)
		if err != nil {
			logger.Errorf("error setting up client %s", err)
			os.Exit(2)
		}
		socketRelayer.c = c
	}
	logger.Infof("Running with %s", socketConfig)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	exitChan := make(chan int)
	go func() {
		for {
			s := <-signalChan
			switch s {
			case syscall.SIGHUP:
				socketRelayer.Drain()
			default:
				socketRelayer.Drain()
				exitChan <- 0
			}
		}
	}()
	go func() {
		code := <-exitChan
		os.Exit(code)
	}()
	runSocketServer(socketRelayer)
}

var socketCmd = &cobra.Command{
	Use:   "socket",
	Short: "A brief description of your command",
	Long:  "",
	Run:   cobrearun,
}

func (s *socketRelay) start() {
	for {
		time.Sleep(s.drainDuration)
		s.Drain()
	}
}

// Drain - Send remaining logs
func (s *socketRelay) Drain() {
	s.mux.Lock()
	defer s.mux.Unlock()
	var (
		err     error
		item    *goque.Item
		bufSize int
		events  []api.Event
	)
	logger.Debugf("Draining socket buffer %d", s.q.Length())
	if s.q.Length() <= 0 {
		return
	}
	events = make([]api.Event, 0)
	for bufSize < maxSize && err == nil {
		item, _ = s.q.Dequeue()
		if item == nil {
			break
		}
		var e api.Event
		bufSize += len(item.Value)
		if bufSize > maxSize {
			break
		}
		err = json.Unmarshal(item.Value, &e)
		if err != nil {
			//TODO handle error
			logger.Errorf("Error decoding  item %b", item.Value)
			return
		}
		events = append(events, e)
		logger.Debugf("Adding event %s\n", e)
	}
	if len(events) <= 0 {
		return
	}

	if resp, err := s.c.RecordBatchEvents(events); err != nil {
		logger.Errorf("error sending batch %s", err)
	} else {
		logger.Debugf("Sent %d events %s", len(events), resp)
	}
}

func (s *socketRelay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("Got request %s %d", r.RemoteAddr, r.ContentLength)
	var body = make([]byte, r.ContentLength)
	n, err := r.Body.Read(body)

	if n == 0 || err != nil {
		if err != io.EOF {
			logger.Errorf("error ready body %d %s %b ", n, err, body)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if err = r.Body.Close(); err != nil {
		logger.Warnf("error closing body buffer %s", err)
	}
	if err = s.send(body); err != nil {
		logger.Warnf("could not send event %s", err)
	}
	w.WriteHeader(202)
	if _, err = w.Write([]byte("ok")); err != nil {
		logger.Warnf("could not set status code %s", err)
	}
}

func (s *socketRelay) send(payload []byte) error {
	_, err := s.q.Enqueue(payload)
	return err
}

func runSocketServer(sr *socketRelay) {
	logger.Infof("Starting Socket server %s ", socketConfig.Socket)
	l, err := net.Listen("unix", socketConfig.Socket)
	if err != nil {
		logger.Errorf("listen error %s", err)
		return
	}
	defer sr.Drain()
	go sr.start()
	logger.Infof("Server ended %s", http.Serve(l, sr))
}

func init() {
	RootCmd.AddCommand(socketCmd)
	socketCmd.PersistentFlags().StringVar(&socketConfig.Server, "server", "http://localhost:8080", "send events to this server")
	socketCmd.PersistentFlags().StringVar(&socketConfig.Socket, "socket", "/tmp/pot.socket", "use this socket")
	socketCmd.PersistentFlags().DurationVar(&socketConfig.Duration, "duration", time.Minute*10, "send events every X minutes (default 5 min)")
	socketCmd.PersistentFlags().BoolVar(&socketConfig.DryRun, "dry-run", false, "don't send events")
}
