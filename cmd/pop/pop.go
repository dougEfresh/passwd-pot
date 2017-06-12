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
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/cmd/listen"
	"github.com/dougEfresh/passwd-pot/cmd/queue"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"github.com/dougEfresh/passwd-pot/log"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

var helloMsg = []byte("+OK POP3 server\r\n")
var okMsg = []byte("+OK\r\n")
var unAuthMsg = []byte("-ERR Password incorrect\r\n")

func (p server) sendEvent(user string, password string, remoteAddrPair []string) {
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
	p.Send(e)
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
	logger.Error("Out of range")
	return "", nil
}

type server struct {
	queue.EventQueue
}

func (p server) HandleConnection(conn net.Conn) {
	defer conn.Close()
	conn.Write(helloMsg)
	reader := bufio.NewReader(conn)
	remoteAddrPair := strings.Split(conn.RemoteAddr().String(), ":")
	var user string
	var password string
	for {
		raw_line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			logger.Errorf("Error reading from client %s", err.Error())
			return
		}
		if err == io.EOF {
			return
		}

		// Parses the command
		cmd, args := getCommand(raw_line)
		logger.Infof("RECV: cmd:%s  args: %s", cmd, args)

		if cmd == "USER" {
			user, _ = getSafeArg(args, 0)
			conn.Write(okMsg)

		} else if cmd == "PASS" {
			password, _ = getSafeArg(args, 0)
			go p.sendEvent(user, password, remoteAddrPair)
			conn.Write(unAuthMsg)

		} else if cmd == "QUIT" {
			return
		} else {
			logger.Warnf("Unknown CMD %s", cmd)
		}
	}
}

func Run(worker work.Worker, l log.Logger) {
	logger = l
	listen.Run(worker, server{
		worker.EventQueue,
	})
}

var logger log.Logger
