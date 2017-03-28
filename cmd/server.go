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
	log "github.com/Sirupsen/logrus"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/gocraft/health"
	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"
)

//Context for http requests
type Context struct {
	eventTransporter
}

var app newrelic.Application

func handlers() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc(getHandler(api.EventURL, handleEvent)).Methods("POST")
	r.HandleFunc(getHandler(api.EventURL, listEvents)).Methods("GET")
	return r
}

func getHandler(path string, h func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	if app != nil {
		return newrelic.WrapHandleFunc(app, path, h)
	} else {
		return path, h
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  "",
	Run:   run,
}

func handleEvent(w http.ResponseWriter, r *http.Request) {
	job := stream.NewJob(fmt.Sprintf("%s", api.EventURL))
	var event Event
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Error reading body %s", err)
		job.EventErr("handle_event_invalid_body", err)
		job.Complete(health.ValidationError)
		return
	}
	if err = json.Unmarshal(b, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("Error unmarshal %s", err)
		job.EventErr("handle_event_invalid_json", err)
		job.Complete(health.ValidationError)
		return
	}

	if event.OriginAddr == "" {
		if r.Header.Get("X-Forwarded-For") != "" {
			log.Debug("Using RemoteAddr from  X-Forwarded-For")
			event.OriginAddr = r.Header.Get("X-Forwarded-For")
		} else {
			//IP:Port
			log.Debugf("Using RemoteAddr as OriginAddr %s", r.RemoteAddr)
			event.OriginAddr = strings.Split(r.RemoteAddr, ":")[0]
		}
	}

	id, err := processEvent(event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Error writing %+v %s", &event, err)
		job.EventErr("handle_event_event_error", err)
		job.Complete(health.Error)
		return
	}
	event.ID = id
	job.Complete(health.Success)
	j, _ := json.Marshal(event)
	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("Content-type", "application/json")
	w.Write(j)

}

func processEvent(event Event) (int64, error) {
	id, resolved, err := defaultEventClient.recordEvent(event)
	if err != nil {
		return 0, err
	}
	event.ID = id
	if !resolved {
		eventChan <- &event
	}
	return id, nil
}

func listEvents(w http.ResponseWriter, r *http.Request) {
	geoEvents := defaultEventClient.list()
	j, _ := json.Marshal(geoEvents)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-type", "application/json")
	w.Write(j)
}

func run(cmd *cobra.Command, args []string) {
	var err error
	setup(cmd, args)
	srv := &http.Server{
		Handler:      handlers(),
		Addr:         config.BindAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Infof("Listing on %s", config.BindAddr)
	//Geo Lookup
	go runLookup()
	err = srv.ListenAndServe()
	if err != nil {
		log.Errorf("Caught error %s", err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().StringVar(&config.Dsn, "dsn", "postgres://postgres:@172.17.0.1/?sslmode=disable", "DSN database url")
	serverCmd.PersistentFlags().StringVar(&config.BindAddr, "bind", "localhost:8080", "bind to this address:port")
	serverCmd.PersistentFlags().StringVar(&config.NewRelic, "new-relic", "", "new relic api key")
	serverCmd.PersistentFlags().BoolVar(&config.NoCache, "no-cache", false, "don't cache geo ip results")
}
