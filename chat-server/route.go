package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	socket "github.com/kwt1326/go-socket-chat-server/chat-server/socket"
)

func route(router *gin.Engine, hub *socket.Hub) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "go chat server"})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong!"})
	})

	router.GET("/ws/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		socket.ServeWs((*socket.Hub)(hub), c.Writer, c.Request, roomId)

		c.JSON(http.StatusOK, gin.H{"message": "webSocket connected!"})
	})

	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	router.Run()
}
