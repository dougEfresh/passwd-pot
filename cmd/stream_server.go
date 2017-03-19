package cmd

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"time"

	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func streamEvents(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "random") {
		serveRandomWs(randomDataHub, w, r)
	} else {
		serveWs(hub, w, r)
	}
}

// serveWs handles websocket requests from the peer.
func serveRandomWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 512)}
	client.hub.register <- client
	go client.writePump()
	client.readPump()
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 512)}
	client.hub.register <- client
	go client.writePump()
	client.readPump()
}

var lastRandomEvent *EventGeo

func startRandomHub(hub *Hub) {
	log.Info("Starting random hub")
	var id int64
	sess := defaultEventClient.db.NewSession(nil)
	whereClause := fmt.Sprintf("id != ? and remote_latitude != ? and remote_longitude != ? and id >= (select max(id) * RANDOM()  from %s) ", eventTable)
	for {
		time.Sleep(1 * time.Second)
		if len(hub.clients) == 0 {
			continue
		}

		if lastRandomEvent == nil {
			if err := sess.Select("max(id)").From(eventGeoTable).LoadValue(&id); err != nil {
				log.Error("Error getting max id")
			}
		} else {
			sess.Select("id").
				From(eventGeoTable).
				Where(whereClause, lastRandomEvent.ID, lastRandomEvent.RemoteLatitude, lastRandomEvent.RemoteLongitude).
				Limit(1).
				OrderBy("id").
				LoadValue(&id)
		}
		if id == 0 {
			log.Error("Could not find an random event!")
		}
		if id != 0 {
			lastRandomEvent = defaultEventClient.broadcastEvent(id, hub)
		}
	}
}