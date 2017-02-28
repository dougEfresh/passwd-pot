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
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"net/http"
	"time"
	"io/ioutil"
	"encoding/json"
	"strings"
)

type Server struct {
	auditClient AuditRecorder
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {

		defaultAuditClient := &AuditClient{
			db: loadDSN(cmd.Flag("dsn").Value.String()),
			geoClient: GeoClientTransporter(geoClient),
		}

		s := &Server{
			auditClient: defaultAuditClient,
		}
		//defer db.Close()
		r := mux.NewRouter()
		r.HandleFunc("/api/v1/audit", s.handleEvent).
			Methods("POST").
			HeadersRegexp("Content-Type", "application/json")
		r.HandleFunc("/api/v1/audit", list).
			Methods("GET")
		srv := &http.Server{
			Handler:      r,
			Addr:         "127.0.0.1:8080",
			WriteTimeout: 3 * time.Second,
			ReadTimeout:  3 * time.Second,
		}
		log.Println("Listing on port 8080")
		log.Fatal(srv.ListenAndServe())
	},
}

func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	var event SshEvent
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if  err = json.Unmarshal(b, &event) ; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//IP:Port
	if event.OriginAddr == "" {
		event.OriginAddr = strings.Split(r.RemoteAddr, ":")[0]
	}

	err = s.auditClient.RecordEvent(&event)
	go s.auditClient.ResolveGeoEvent(&event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error writing %+v %s", &event, err)
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
