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
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/service"
	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/trace"
	"strings"
	"time"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  "",
	Run:   run,
}
var geoCache *Cache = NewCache()
var eventClient *service.EventClient
var resolveClient *service.ResolveClient
var app newrelic.Application

func handlers() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc(getHandler(api.EventURL, handleEvent)).Methods("POST")
	//r.HandleFunc(getHandler(api.EventURL, listEvents)).Methods("GET")
	return r
}

func getHandler(path string, h func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	if app != nil {
		return newrelic.WrapHandleFunc(app, path, h)
	} else {
		return path, h
	}
}

func handleEvent(w http.ResponseWriter, r *http.Request) {
	var event api.Event
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Errorf("Error reading body %s", err)
		return
	}
	if err = json.Unmarshal(b, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Errorf("Error unmarshal %s", err)
		return
	}

	if event.OriginAddr == "" {
		if r.Header.Get("X-Forwarded-For") != "" {
			logger.Debug("Using RemoteAddr from  X-Forwarded-For")
			event.OriginAddr = r.Header.Get("X-Forwarded-For")
		} else {
			//IP:Port
			logger.Debugf("Using RemoteAddr as OriginAddr %s", r.RemoteAddr)
			event.OriginAddr = strings.Split(r.RemoteAddr, ":")[0]
		}
	}
	var l prometheus.Labels
	var ok bool
	if l, ok = labels[event.OriginAddr]; !ok {
		l = prometheus.Labels{"origin": event.OriginAddr}
	}
	recordCounter.With(l).Inc()
	id, err := eventClient.RecordEvent(event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Errorf("Error writing %+v %s", &event, err)
		return
	}
	event.ID = id
	go resolveEvent(event)
	j, _ := json.Marshal(id)
	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("Content-Type", "application/json")
	w.Write(j)

}

func resolveEvent(event api.Event) {
	var err error
	var ids []int64
	rId, _ := geoCache.get(event.RemoteAddr)
	oId, _ := geoCache.get(event.OriginAddr)
	if rId > 0 && oId > 0 {
		e := resolveClient.MarkEvent(event.ID, rId, true)
		if e != nil {
			err = e
		}
		e = resolveClient.MarkEvent(event.ID, oId, false)
		if e != nil {
			err = e
		}
	} else {
		ids, err = resolveClient.ResolveEvent(event)
		if err != nil {
			geoCache.set(event.RemoteAddr, ids[0])
			geoCache.set(event.OriginAddr, ids[1])
		}
	}
	if err != nil {
		logger.Errorf("Error looking up %s %s", event, err)
	}
}

func run(cmd *cobra.Command, args []string) {
	setupLogger(cmd.Name())
	setup(cmd, args)
	srv := &http.Server{
		Handler:      handlers(),
		Addr:         config.BindAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	db := loadDSN(config.Dsn)
	eventClient, _ = service.NewEventClient(service.SetEventLogger(logger), service.SetEventDb(db))
	resolveClient, _ = service.NewResolveClient(service.SetResolveLogger(logger), service.SetResolveDb(db))
	if config.Trace {
		logger.Info("Enabling trace")
		f, _ := os.Create(fmt.Sprintf("/tmp/trace-%s.out", cmd.Name()))
		defer f.Close()
		err := trace.Start(f)
		if err != nil {
			panic(err)
		}
		defer trace.Stop()
	}
	err := srv.ListenAndServe()
	if err != nil {
		logger.Errorf("Caught error %s", err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().StringVar(&config.Dsn, "dsn", "postgres://postgres:@172.17.0.1/?sslmode=disable", "DSN database url")
	serverCmd.PersistentFlags().StringVar(&config.BindAddr, "bind", "localhost:8080", "bind to this address:port")
	serverCmd.PersistentFlags().StringVar(&config.NewRelic, "new-relic", "", "new relic api key")
	serverCmd.PersistentFlags().BoolVar(&config.NoCache, "no-cache", false, "don't cache geo ip results")
	serverCmd.PersistentFlags().BoolVar(&config.Trace, "trace", false, "enable trace")
	prometheus.MustRegister(recordCounter)
}

var labels = make(map[string]prometheus.Labels)
var recordCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "passwdpot",
	Name:      "record",
	Help:      "count of requests",
	Subsystem: "total",
}, []string{"origin"})
