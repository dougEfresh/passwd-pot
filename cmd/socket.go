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
	"github.com/cenkalti/backoff"
	"github.com/dougEfresh/passwd-pot/cmd/listen"
	"github.com/dougEfresh/passwd-pot/cmd/log"
	"github.com/spf13/cobra"
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
	if logger.GetLevel() == log.DebugLevel {
		logger.Debugf("Response %s", string(resp[:bytes.IndexByte(resp, 0)]))
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
		setupLogger(cmd.Name())
		if config.Pprof != "" {
			go func() { logger.Error(http.ListenAndServe(config.Pprof, nil)) }()
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
		logger.Errorf("Error reading %s", err)
	}
	go sendEvent(request[:bytes.IndexByte(request, 0)], sr)
	c.Write(socketResponse)
}

func sendEvent(request []byte, sr socketRelayer) {
	if logger.GetLevel() == log.DebugLevel {
		logger.Debugf("Socket Request %s", string(request))
	}
	err := backoff.Retry(func() error {
		return sr.Send(request)
	}, backoff.NewExponentialBackOff())

	if err != nil {
		logger.Errorf("Error sending %s (%s)", string(request), err)
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
		logger.Errorf("listen error %s", err)
		return
	}
	lc := make(chan net.Conn, 100)
	go listen.AcceptConnection(l, lc)
	defer func() {
		l.Close()
		os.Remove(socketConfig.Socket)
	}()
	for {
		select {
		case conn := <-lc:
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
