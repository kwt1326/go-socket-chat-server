package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	cons "local/constants"
)

func (c *Subscription) readPump() {
	hub := c.ClientP.HubP
	conn := c.ClientP.ConnP

	defer func() {
		hub.Unregister <- (cons.Subscription)(*c)
		conn.Close()
	}()

	conn.SetReadLimit(cons.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(cons.PongWait))
	conn.SetPongHandler(func(string) error { c.ClientP.ConnP.SetReadDeadline(time.Now().Add(cons.PongWait)); return nil })

	for {
		_, message, err := c.ClientP.ConnP.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, cons.Newline, cons.Space, -1))
		hub.Broadcast <- (cons.Message)(Message{Value: message, RoomId: c.RoomId})
	}
}

func (c *Subscription) writePump() {
	ticker := time.NewTicker(cons.PingPeriod)
	client := c.ClientP
	conn := c.ClientP.ConnP

	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.SendChan:
			conn.SetWriteDeadline(time.Now().Add(cons.WriteWait))
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
				w.Write(cons.Newline)
				w.Write(<-client.SendChan)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(cons.WriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *cons.Hub, writer http.ResponseWriter, request *http.Request, roomId string) {
	conn, err := cons.Upgrader.Upgrade(writer, request, nil)

	if err != nil {
		log.Println(err)
		return
	}

	subscription := &Subscription{ClientP: &cons.Client{
		HubP: hub, ConnP: conn, SendChan: make(chan []byte, 256),
	}, RoomId: roomId}

	hub.Register <- (cons.Subscription)(*subscription)

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go subscription.writePump()
	go subscription.readPump()
}
