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
	"context"
	"encoding/json"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent"
	"net/http"
	"strings"
)

var app newrelic.Application

func MakeHTTPHandler(s EventService, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path(api.EventURL).Handler(getHandle(api.EventURL, httptransport.NewServer(e.RecordEndpoint, decodeEvent, encodeResponse, options...)))
	return r
}

func getHandle(path string, h http.Handler) http.Handler {
	var nh http.Handler
	if app == nil {
		return h
	}
	_, nh = newrelic.WrapHandle(app, path, h)
	return nh
}

func decodeEvent(_ context.Context, r *http.Request) (request interface{}, err error) {
	var event Event
	if e := json.NewDecoder(r.Body).Decode(&event); e != nil {
		return nil, e
	}
	if event.OriginAddr == "" {
		if r.Header.Get("X-Forwarded-For") != "" {
			logger.Log("msg", "Using RemoteAddr from  X-Forwarded-For")
			event.OriginAddr = r.Header.Get("X-Forwarded-For")
		} else {
			//IP:Port
			logger.Log("msg", "Using RemoteAddr as OriginAddr ")
			event.OriginAddr = strings.Split(r.RemoteAddr, ":")[0]
		}
	}
	return event, nil
}

type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}
