package socket

import (
	"log"
)

type Message struct {
	Value  []byte
	RoomId string
}

// manage connections, broadcast messages
type Hub struct {
	// Registered clients.
	Rooms map[string]map[*Client]bool // chat room roomId(key):value(map[*Client]bool) 로 매핑

	// Inbound messages from the clients.
	Broadcast chan Message

	// Register requests from the clients.
	Register chan Subscription

	// Unregister requests from clients.
	Unregister chan Subscription
}

// Hub 는 지속적으로 채널을 통해 처리하는 역할을 한다.
func (h *Hub) Run() {
	for { // while(true)
		select {

		// Channel Register
		case subscription := <-h.Register:
			clients := h.Rooms[subscription.RoomId]

			if clients == nil {
				clients = make(map[*Client]bool)
				h.Rooms[subscription.RoomId] = clients

				log.Println("New Room Created:", subscription.RoomId)
			}
			// add client to room
			h.Rooms[subscription.RoomId][subscription.ClientP] = true

		// Channel Unregister
		case subscription := <-h.Unregister:
			clients := h.Rooms[subscription.RoomId]
			client := subscription.ClientP

			if clients != nil && client != nil {
				if _, ok := clients[client]; ok {
					CloseRoom(h, client, clients, subscription.RoomId)
					delete(clients, client)
					close(client.SendChan)
					if len(clients) == 0 {
						delete(h.Rooms, subscription.RoomId)
					}
				}
			}

		// Channel Message Broadcast
		case message := <-h.Broadcast:
			clients := h.Rooms[message.RoomId]

			for client := range clients {
				select {
				case client.SendChan <- message.Value:
				default:
					CloseRoom(h, client, clients, message.RoomId)
				}
			}
		}
	}
}

func CloseRoom(h *Hub, client *Client, clients map[*Client]bool, roomId string) {
	close(client.SendChan)
	delete(h.Rooms[roomId], client)
	if len(clients) == 0 {
		delete(h.Rooms, roomId)
	}
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan Message),
		Register:   make(chan Subscription),
		Unregister: make(chan Subscription),
		Rooms:      make(map[string]map[*Client]bool),
	}
}
