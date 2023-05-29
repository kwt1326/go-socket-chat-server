package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"

	socket "github.com/kwt1326/go-socket-chat-server/chat-server/socket"
)

func initHub() *socket.Hub {
	hub := socket.NewHub()
	go hub.Run()
	return hub
}

func run(hub *socket.Hub) {
	router := gin.New()
	route(router, hub)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	run(initHub())
}
