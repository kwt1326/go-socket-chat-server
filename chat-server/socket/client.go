package socket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Hub and Connection Middle Relationship Object
type Client struct {
	HubP     *Hub
	ConnP    *websocket.Conn
	SendChan chan []byte // Buffered channel of outbound messages.
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, writer http.ResponseWriter, request *http.Request, roomId string) {
	conn, err := Upgrader.Upgrade(writer, request, nil)

	if err != nil {
		log.Println(err)
		return
	}

	subscription := &Subscription{ClientP: &Client{
		HubP: hub, ConnP: conn, SendChan: make(chan []byte, 256),
	}, RoomId: roomId}

	hub.Register <- (Subscription)(*subscription)

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go subscription.writePump()
	go subscription.readPump()
}
