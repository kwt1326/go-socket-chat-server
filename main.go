package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	cons "local/constants"
)

// main define
type Hub cons.Hub
type Client cons.Client
type Message cons.Message
type Subscription cons.Subscription

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func initHub() *Hub {
	hub := newHub()
	go hub.run()
	return hub
}

func initAndRunServer(hub *Hub) {
	router := gin.New()
	router.LoadHTMLFiles("index.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong!"})
	})

	router.GET("/ws/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		serveWs((*cons.Hub)(hub), c.Writer, c.Request, roomId)

		c.JSON(http.StatusOK, gin.H{"message": "webSocket connected!"})
	})

	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	initAndRunServer(initHub())
}
