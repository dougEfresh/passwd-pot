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

/*
import (
	"fmt"
	"github.com/dougEfresh/passwd-pot/api"
	"github.com/dougEfresh/passwd-pot/log"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"time"
)

// streamCmd represents the stream command
var streamingEndpoint string

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		streamEvents()
	},
}

func streamEvents() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	c, resp, err := websocket.DefaultDialer.Dial(streamingEndpoint, nil)
	if err != nil {
		log.Errorf("websocket.New Error: %s\nResp:%+v", err, resp)
		return
	}

	defer c.Close()
	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

func init() {
	RootCmd.AddCommand(streamCmd)
	streamCmd.Flags().StringVarP(&streamingEndpoint, "server", "s",
		fmt.Sprintf("ws://localhost:8080%s/random", api.StreamURL),
		"server endpoint")
}

 */
