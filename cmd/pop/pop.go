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

// props to https://github.com/r0stig/golang-pop3/blob/master/main.go

package pop

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

var helloMsg = []byte("+OK POP3 server\r\n")
var okMsg = []byte("+OK\r\n")
var unAuthMsg = []byte("-ERR Password incorrect\r\n")

func sendEvent(user string, password string, remoteAddrPair []string, worker work.Worker) {
	remotePort, err := strconv.Atoi(remoteAddrPair[1])
	if err != nil {
		remotePort = 0
	}
	e := &api.Event{
		User:        user,
		Passwd:      password,
		Time:        api.EventTime(time.Now().UTC()),
		RemoteName:  remoteAddrPair[0],
		RemoteAddr:  remoteAddrPair[0],
		RemotePort:  remotePort,
		Application: "pop-passwd-pot",
		Protocol:    "pop",
	}

	if names, err := net.LookupAddr(e.RemoteAddr); err == nil && len(names) > 0 {
		e.RemoteName = names[0]
	}
	worker.EventQueue.Send(e)
}

func handleClient(conn net.Conn, worker work.Worker) {
	defer conn.Close()
	conn.Write(helloMsg)
	reader := bufio.NewReader(conn)
	remoteAddrPair := strings.Split(conn.RemoteAddr().String(), ":")
	var user string
	var password string
	for {
		raw_line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Errorf("Error reading from client %s", err.Error())
			return
		}
		if err == io.EOF {
			return
		}

		// Parses the command
		cmd, args := getCommand(raw_line)
		log.Debugf("RECV: cmd:%s  args: %s", cmd, args)

		if cmd == "USER" {
			user, _ = getSafeArg(args, 0)
			conn.Write(okMsg)

		} else if cmd == "PASS" {
			password, _ = getSafeArg(args, 0)
			go sendEvent(user, password, remoteAddrPair, worker)
			conn.Write(unAuthMsg)

		} else if cmd == "QUIT" {
			return
		} else {
			log.Warnf("Unknown CMD %s", cmd)
		}
	}
}

func getCommand(line string) (string, []string) {
	line = strings.Trim(line, "\r \n")
	cmd := strings.Split(line, " ")
	return cmd[0], cmd[1:]
}

func getSafeArg(args []string, nr int) (string, error) {
	if nr < len(args) {
		return args[nr], nil
	}
	log.Error("Out of range")
	return "", nil
}

func Run(worker work.Worker) {
	defer worker.Wg.Done()
	if worker.Addr == "" {
		log.Warn("Not starting pop pot")
		return
	}
	ln, err := net.Listen("tcp", worker.Addr)
	if err != nil {
		log.Errorf("Cannot bind to %s %s", worker.Addr, err)
		return
	}
	log.Infof("Started pop pot on %s", worker.Addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		// run as goroutine
		go handleClient(conn, worker)
	}
}
