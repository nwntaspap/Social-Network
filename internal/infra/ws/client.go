package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	UserID     string
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	OpenChatId string
}

func (c *Client) Send(msg []byte) {
	c.send <- msg
}

func NewClient(userID string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		UserID: userID,
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
}

// ReadPump pumps messages from webSocket to the hub.
// Must run in its own goroutine.
func (c *Client) ReadPump(onMessage func(client *Client, msg []byte)) {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()
	defer recoverWS()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws read error: %v", err)
			}
			break
		}
		onMessage(c, msg)
	}
}

// WritePump pumps messages from the send channel to the WebSocket.
// Must run in its own goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.hub.Unregister(c)
		ticker.Stop()
		_ = c.conn.Close()
	}()
	defer recoverWS()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}

func recoverWS() {
	if r := recover(); r != nil {
		log.Printf("ws panic: %v", r)
	}
}
