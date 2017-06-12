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

package http

import (
	"encoding/base64"
	"net/http"
	"strings"

	"time"

	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/queue"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"github.com/dougEfresh/passwd-pot/log"
	"net"
	"strconv"
)

var unAuthorieds []byte = []byte("401 Unauthorized\n")

func sendEvent(user string, password string, r *http.Request, p *potHttpHandler) {
	logger.Debugf("processing request %s %s", user, password)
	remoteAddrPair := strings.Split(r.RemoteAddr, ":")
	remotePort, err := strconv.Atoi(remoteAddrPair[1])
	if err != nil {
		remotePort = 0
	}
	e := &api.Event{
		Time:          api.EventTime(time.Now().UTC()),
		User:          user,
		Passwd:        password,
		RemoteAddr:    remoteAddrPair[0],
		RemoteName:    remoteAddrPair[0],
		RemotePort:    remotePort,
		Application:   "http-passwd-pot",
		Protocol:      "http-basic-auth",
		RemoteVersion: r.UserAgent(),
	}

	if r.Header.Get("X-Forwarded-For") != "" {
		logger.Debug("Using RemoteAddr from X-Forwarded-For")
		e.RemoteAddr = r.Header.Get("X-Forwarded-For")
		e.RemoteName = e.RemoteAddr
	}
	if names, err := net.LookupAddr(e.RemoteAddr); err == nil && len(names) > 0 {
		e.RemoteName = names[0]
	}
	p.eventQueue.Send(e)
}

func processAuth(r *http.Request, p *potHttpHandler) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	logger.Debugf("HEADER %s", s)
	if len(s) != 2 {
		return
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		logger.Errorf("Error decoding %s %s", s[1], err)
		return
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return
	}

	go sendEvent(pair[0], pair[1], r, p)
}

func handleRequest(w http.ResponseWriter, r *http.Request, p *potHttpHandler) {
	processAuth(r, p)
	w.Header().Set("WWW-Authenticate", `Basic realm="default"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(unAuthorieds)
}

type potHttpHandler struct {
	eventQueue queue.EventQueue
}

func (p *potHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, p)
}

// Run the worker
func Run(worker work.Worker, l log.Logger) {
	logger = l
	defer worker.Wg.Done()
	if worker.Addr == "" {
		logger.Warn("Not starting http pot")
		return
	}
	srv := &http.Server{
		Handler: &potHttpHandler{
			eventQueue: worker.EventQueue,
		},
		Addr:         worker.Addr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	logger.Infof("Started http pot on %s", worker.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Errorf("Error starting server %v", err)
	}
}

var logger log.Logger
