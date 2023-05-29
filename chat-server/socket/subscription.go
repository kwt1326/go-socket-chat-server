package socket

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"

	constants "github.com/kwt1326/go-socket-chat-server/chat-server/constants"
)

type Subscription struct {
	ClientP *Client
	RoomId  string
}

func (c *Subscription) readPump() {
	hub := c.ClientP.HubP
	conn := c.ClientP.ConnP

	defer func() {
		hub.Unregister <- (Subscription)(*c)
		conn.Close()
	}()

	conn.SetReadLimit(constants.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(constants.PongWait))
	conn.SetPongHandler(func(string) error { c.ClientP.ConnP.SetReadDeadline(time.Now().Add(constants.PongWait)); return nil })

	for {
		_, message, err := c.ClientP.ConnP.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, constants.Newline, constants.Space, -1))
		hub.Broadcast <- (Message)(Message{Value: message, RoomId: c.RoomId})
	}
}

func (c *Subscription) writePump() {
	ticker := time.NewTicker(constants.PingPeriod)
	client := c.ClientP
	conn := c.ClientP.ConnP

	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.SendChan:
			conn.SetWriteDeadline(time.Now().Add(constants.WriteWait))
			if !ok {
				// The hub closed the channel.
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(client.SendChan)
			for i := 0; i < n; i++ {
				w.Write(constants.Newline)
				w.Write(<-client.SendChan)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(constants.WriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}