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
	"crypto/tls"
	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/cenkalti/backoff"
	"github.com/spf13/cobra"
	"log/syslog"
	"net"
	"net/http"
	"os"
	"sync"
)

var socketResponse []byte = []byte("HTTP/1.1 202 Accepted\r\n\r\n")
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 2038)
	},
}

type socketRelayer interface {
	Send(r []byte) error
}

type socketRelay struct {
}

func (s socketRelay) Send(r []byte) error {
	if socketConfig.DryRun {
		return nil
	}
	conn, err := tls.Dial("tcp", socketConfig.Server, &tls.Config{})
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(r)
	if err != nil {
		return err
	}
	resp := bufferPool.Get().([]byte)
	defer bufferPool.Put(resp)
	_, err = conn.Read(resp)
	if err != nil {
		return err
	}
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("Response %s", string(resp[:bytes.IndexByte(resp, 0)]))
	}
	return nil
}

var socketConfig struct {
	Pprof  string
	Server string
	Socket string
	DryRun bool
}

var socketCmd = &cobra.Command{
	Use:   "socket",
	Short: "A brief description of your command",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("starting %s %v", cmd.Name(), socketConfig)
		if config.Debug {
			log.SetLevel(log.DebugLevel)
		}
		if config.Syslog != "" {
			if syslogHook, err := logrus_syslog.NewSyslogHook("tcp", config.Syslog, syslog.LOG_LOCAL0, "passwd-potter"); err != nil {
				log.Error("Unable to connect to local syslog daemon")
			} else {
				log.AddHook(syslogHook)
			}
		}
		if config.Pprof != "" {
			go func() { log.Error(http.ListenAndServe(config.Pprof, nil)) }()
		}
		runSocketServer(socketRelay{})
	},
}

func handleSocketRequest(c net.Conn, sr socketRelayer) {
	defer c.Close()
	request := bufferPool.Get().([]byte)
	defer bufferPool.Put(request)
	_, err := c.Read(request)
	if err != nil {
		log.Errorf("Error reading %s", err)
	}
	go sendEvent(request[:bytes.IndexByte(request, 0)], sr)
	c.Write(socketResponse)
}

func sendEvent(request []byte, sr socketRelayer) {
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("Socket Request %s", string(request))
	}
	err := backoff.Retry(func() error {
		return sr.Send(request)
	}, backoff.NewExponentialBackOff())

	if err != nil {
		log.Errorf("Error sending %s (%s)", string(request), err)
	}
}

func acceptConnection(listener net.Listener, listen chan<- net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		listen <- conn
	}
}

func runSocketServer(sr socketRelayer) {

	f, err := os.Open(socketConfig.Socket)
	if err != nil {
		f.Close()
		os.Remove(socketConfig.Socket)
	}
	l, err := net.Listen("unix", socketConfig.Socket)
	if err != nil {
		log.Errorf("listen error %s", err)
		return
	}
	listen := make(chan net.Conn, 100)
	go acceptConnection(l, listen)
	defer func() {
		l.Close()
		os.Remove(socketConfig.Socket)
	}()
	for {
		select {
		case conn := <-listen:
			go handleSocketRequest(conn, sr)
		}
	}

}

func init() {
	RootCmd.AddCommand(socketCmd)
	socketCmd.PersistentFlags().StringVar(&socketConfig.Server, "server", "http://localhost:8080", "send events to this server")
	socketCmd.PersistentFlags().StringVar(&socketConfig.Socket, "socket", "/tmp/pot.socket", "use this socket")
	socketCmd.PersistentFlags().BoolVar(&socketConfig.DryRun, "dry-run", false, "don't send events")
	for i := 0; i < 100; i++ {
		bufferPool.Put(make([]byte, 2048))
	}
}
