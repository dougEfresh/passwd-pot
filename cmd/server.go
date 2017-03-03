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
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"log/syslog"
)

type server struct {
	auditClient auditRecorder
}

const (
	auditEventURL = "/api/v1/audit"
)

func handlers(s *server) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(auditEventURL, s.handleEvent).
		Methods("POST").
		HeadersRegexp("Content-Type", "application/json")

	return router
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if config.Debug {
			log.SetLevel(log.DebugLevel)
		}
		defaultAuditClient := &auditClient{
			db:        loadDSN(config.Dsn),
			geoClient: geoClientTransporter(geoClient),
		}

		s := &server{
			auditClient: defaultAuditClient,
		}
		srv := &http.Server{
			Handler:      handlers(s),
			Addr:         config.BindAddr,
			WriteTimeout: 3 * time.Second,
			ReadTimeout:  3 * time.Second,
		}
		if config.Syslog != "" {
			hook, err := logrus_syslog.NewSyslogHook("tcp", "localhost:514", syslog.LOG_INFO, "ssh-password-pot")
			if err != nil {
				log.Error("Unable to connect to local syslog daemon")
			} else {
				log.AddHook(hook)
			}
		}
		log.Infof("Listing on %s", config.BindAddr)
		log.Fatal(srv.ListenAndServe())
	},
}

func (s *server) handleEvent(w http.ResponseWriter, r *http.Request) {
	var event SSHEvent
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
		return
	}
	if err = json.Unmarshal(b, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("Error reading %s", err)
		return
	}
	//IP:Port
	if event.OriginAddr == "" {
		if r.Header.Get("X-Forwarded-For") != "" {
			event.OriginAddr = r.Header.Get("X-Forwarded-For")
		} else {
			event.OriginAddr = strings.Split(r.RemoteAddr, ":")[0]
		}
		log.Debugf("Using RemoteAddr as OriginAddr %s", r.RemoteAddr)
	}

	err = s.auditClient.recordEvent(&event)
	go s.auditClient.resolveGeoEvent(&event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Error writing %+v %s", &event, err)
		return
	}

	j, _ := json.Marshal(event)
	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("Content-type", "application/json")
	w.Write(j)
}

func list(w http.ResponseWriter, r *http.Request) {
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
