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

package listen

import (
	"github.com/dougEfresh/passwd-pot/cmd/log"
	"github.com/dougEfresh/passwd-pot/cmd/work"
	"net"
)

type Handler interface {
	HandleConnection(c net.Conn)
}

func AcceptConnection(listener net.Listener, listen chan<- net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf("Error getting connection %s", err)
			continue
		}
		listen <- conn
	}
}

func Run(worker work.Worker, h Handler) {
	defer worker.Wg.Done()
	if worker.Addr == "" {
		logger.Warnf("Not starting %s pot", worker.Name)
		return
	}
	ln, err := net.Listen("tcp", worker.Addr)
	if err != nil {
		logger.Errorf("Cannot bind to %s %s", worker.Addr, err)
		return
	}
	logger.Infof("Started pot %s on %s", worker.Name, worker.Addr)
	lc := make(chan net.Conn)
	go AcceptConnection(ln, lc)
	defer ln.Close()
	for {
		select {
		case conn := <-lc:
			go h.HandleConnection(conn)
		}
	}
}

var logger log.Logger
