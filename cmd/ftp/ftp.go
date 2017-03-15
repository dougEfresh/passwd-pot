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

package ftp

import (
	log "github.com/Sirupsen/logrus"
	"strings"

	"time"

	"bufio"
	"errors"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/queue"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"net"
	"strconv"
)

var unAuthorized []byte = []byte("530 Login authentication failed\r\n")
var userOk []byte = []byte("331 User OK\r\n")

func (p *potHandler) sendEvent(user string, password string, conn net.Conn) {
	log.Debugf("processing request %s %s", user, password)
	remoteAddrPair := strings.Split(conn.RemoteAddr().String(), ":")
	remotePort, err := strconv.Atoi(remoteAddrPair[1])
	if err != nil {
		remotePort = 0
	}
	e := &api.Event{
		Time:        api.EventTime(time.Now().UTC()),
		User:        user,
		Passwd:      password,
		RemoteAddr:  remoteAddrPair[0],
		RemoteName:  remoteAddrPair[0],
		RemotePort:  remotePort,
		Application: "ftp-passwd-pot",
		Protocol:    "ftp",
	}

	if names, err := net.LookupAddr(e.RemoteAddr); err == nil && len(names) > 0 {
		e.RemoteName = names[0]
	}
	p.eventQueue.Send(e)
}

type potHandler struct {
	eventQueue queue.EventQueue
}

func (p *potHandler) handleConnection(conn net.Conn) {
	var user string
	var pass string
	defer conn.Close()
	if _, err := conn.Write([]byte("220 This is a private system - No anonymous login\r\n")); err != nil {
		log.Errorf("Error sending 220 %s", err)
		return
	}
	commandPair, err := readCommand(conn)
	if err != nil {
		log.Errorf("Error reading cmd %s", err)
		return
	}
	if len(commandPair) >= 2 {
		user = commandPair[1]
	}
	if _, err := conn.Write(userOk); err != nil {
		log.Error("Error writing 331 User")
		return
	}
	commandPair, err = readCommand(conn)
	if err != nil {
		log.Errorf("Error reading command %s", err)
		return
	}
	if len(commandPair) >= 2 {
		pass = commandPair[1]
	}
	p.sendEvent(user, pass, conn)
	conn.Write(unAuthorized)
}

func readCommand(conn net.Conn) ([]string, error) {
	c, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	c = strings.Replace(c, "\r\n", "", 1)
	log.Debugf("CMD: %s", c)
	if strings.Contains(c, "QUIT") {
		return nil, errors.New("QUIT issued")
	}
	return strings.Split(c, " "), nil
}

func Run(worker *work.Worker) {
	defer worker.Wg.Done()
	if worker.Addr == "" {
		log.Warn("Not starting ftp pot")
		return
	}
	ln, err := net.Listen("tcp", worker.Addr)
	if err != nil {
		log.Errorf("Cannot bind to %s %s", worker.Addr, err)
		return
	}
	log.Infof("Started ftp pot on %s", worker.Addr)
	p := &potHandler{
		eventQueue: worker.EventQueue,
	}
	for {
		conn, _ := ln.Accept()
		go p.handleConnection(conn)
	}
}
