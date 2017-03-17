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

package telnet

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

var welcomMsg = []byte("Welcome. Please enter your loging credentials.\r\n")
var loginPrompt = []byte("login: ")
var passPrompt = []byte("password: ")
var unauthorizedMsg = []byte("Unauthorized\r\n")

type potHandler struct {
	worker work.Worker
}

func (h potHandler) sendEvent(user string, password string, remoteAddrPair []string) {
	remotePort, err := strconv.Atoi(remoteAddrPair[1])
	if err != nil {
		remotePort = 0
	}
	e := &api.Event{
		User:        user,
		Passwd:      password,
		Time:        api.EventTime(time.Now().UTC()),
		RemoteAddr:  remoteAddrPair[0],
		RemoteName:  remoteAddrPair[0],
		RemotePort:  remotePort,
		Application: "telnet-passwd-pot",
		Protocol:    "telnet",
	}
	h.worker.EventQueue.Send(e)
}

func readLine(conn net.Conn) (string, error) {
	c, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		if err != io.EOF {
			log.Errorf("Error reading %s", err)
		}
		return "", err
	}
	return strings.Trim(c, "\r \n"), nil
}

func (h potHandler) handleNewConnection(conn net.Conn) {
	var cnt int
	var user string
	remoteAddrPair := strings.Split(conn.RemoteAddr().String(), ":")
	log.Debugf("Got new connection from %s", conn.RemoteAddr())
	defer conn.Close()
	conn.Write(welcomMsg)
	if _, err := conn.Write(loginPrompt); nil != err {
		log.Errorf("Problem long writing welcome message: %v", err)
		return
	}
	for {
		line, err := readLine(conn)
		if err != nil {
			if err != io.EOF {
				log.Errorf("error reading line %s", err)
			}
			return
		}
		cnt++
		log.Debug("recv ", line)
		if cnt%2 == 1 {
			user = line
			if _, err := conn.Write(passPrompt); nil != err {
				if err != io.EOF {
					log.Errorf("error reading line %s", err)
				}
				return
			}
		} else {
			go h.sendEvent(user, line, remoteAddrPair)
			conn.Write(unauthorizedMsg)
			if _, err := conn.Write(loginPrompt); nil != err {
				if err != io.EOF {
					log.Errorf("error reading line %s", err)
				}
				return
			}
		}
		continue
	}
}

func Run(worker work.Worker) {
	handler := potHandler{
		worker: worker,
	}
	defer worker.Wg.Done()
	if worker.Addr == "" {
		log.Warnf("Not starting telnet pot")
		return
	}
	log.Infof("Started telnet pot on %s", worker.Addr)
	ln, err := net.Listen("tcp", worker.Addr)
	if nil != err {
		log.Errorf("Error on %s %s", worker.Addr, err)
	}
	for {
		conn, _ := ln.Accept()
		go handler.handleNewConnection(conn)
	}
}
