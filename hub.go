package main

import (
	"log"

	cons "local/constants"
)

func newHub() *Hub {
	return &Hub{
		Broadcast:  make(chan cons.Message),
		Register:   make(chan cons.Subscription),
		Unregister: make(chan cons.Subscription),
		Rooms:      make(map[string]map[*cons.Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case subscription := <-h.Register:
			clients := h.Rooms[subscription.RoomId]

			if clients == nil {
				clients = make(map[*cons.Client]bool)
				h.Rooms[subscription.RoomId] = clients

				log.Println("New Room Created:", subscription.RoomId)
			}
			// add client to room
			h.Rooms[subscription.RoomId][subscription.ClientP] = true
		case subscription := <-h.Unregister:
			clients := h.Rooms[subscription.RoomId]
			client := subscription.ClientP

			if clients != nil && client != nil {
				if _, ok := clients[client]; ok {
					closeRoom(h, client, clients, subscription.RoomId)
					delete(clients, client)
					close(client.SendChan)
					if len(clients) == 0 {
						delete(h.Rooms, subscription.RoomId)
					}
				}
			}
		case message := <-h.Broadcast:
			clients := h.Rooms[message.RoomId]

			for client := range clients {
				select {
				case client.SendChan <- message.Value:
				default:
					closeRoom(h, client, clients, message.RoomId)
				}
			}
		}
	}
}

func closeRoom(h *Hub, client *cons.Client, clients map[*cons.Client]bool, roomId string) {
	close(client.SendChan)
	delete(h.Rooms[roomId], client)
	if len(clients) == 0 {
		delete(h.Rooms, roomId)
	}
}
