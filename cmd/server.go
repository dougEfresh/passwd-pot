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
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/trace"
	"strings"
	"time"

	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/service"
	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  "",
	Run:   run,
}
var geoCache *Cache = NewCache()
var eventClient api.Transporter
var resolveClient *service.ResolveClient
var app newrelic.Application

func handlers() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc(getHandler(api.EventURL, handleEvent)).Methods("POST")
	r.HandleFunc(getHandler(api.EventCountryStatsUrl, handleEventCountryStats)).Methods("GET")
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

func handleEventCountryStats(w http.ResponseWriter, _ *http.Request) {
	cached, found := ch.Get("cc_stats")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	if found {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(cached.([]byte))
		return
	}
	stats, err := eventClient.GetCountryStats()
	if err != nil {
		logger.Errorf("Error getting stats %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(stats)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
	ch.Set("cc_stats", b, cache.DefaultExpiration)
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
	//TODO
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
	var err error
	if len(config.GeoDB) > 0 {
		resolveClient, err = service.NewResolveClient(service.SetResolveLogger(logger), service.SetResolveDb(db), service.SetGeoDb(config.GeoDB))
		if err != nil {
			logger.Errorf("Cannot open geo db %s", err)
			os.Exit(-1)
		}
	} else {
		resolveClient, _ = service.NewResolveClient(service.SetResolveLogger(logger), service.SetResolveDb(db))
	}
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
	err = srv.ListenAndServe()
	if err != nil {
		logger.Errorf("Caught error %s", err)
		os.Exit(-1)
	}
}

var ch *cache.Cache = cache.New(10*time.Minute, 20*time.Minute)

func init() {
	RootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().StringVar(&config.Dsn, "dsn", "postgres://postgres:@172.17.0.1/?sslmode=disable", "DSN database url")
	serverCmd.PersistentFlags().StringVar(&config.BindAddr, "bind", "localhost:8080", "bind to this address:port")
	serverCmd.PersistentFlags().StringVar(&config.NewRelic, "new-relic", "", "new relic api key")
	serverCmd.PersistentFlags().BoolVar(&config.NoCache, "no-cache", false, "don't cache geo ip results")
	serverCmd.PersistentFlags().BoolVar(&config.Trace, "trace", false, "enable trace")
	serverCmd.PersistentFlags().StringVar(&config.GeoDB, "geo-db", "", "location of geoLite2 db")
	prometheus.MustRegister(recordCounter)
}

var labels = make(map[string]prometheus.Labels)
var recordCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "passwdpot",
	Name:      "record",
	Help:      "count of requests",
	Subsystem: "total",
}, []string{"origin"})
