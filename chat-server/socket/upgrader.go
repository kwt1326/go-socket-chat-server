package socket

import (
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	cons "github.com/kwt1326/go-socket-chat-server/chat-server/constants"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("origin")
		origins := func() []string {
			if os.Getenv("GIN_MODE") == "debug" {
				return cons.DevOrigins
			} else {
				return cons.ProdOrigins
			}
		}()
		for _, allowOrigin := range origins {
			if origin == allowOrigin {
				return true
			}
		}
		return true // TODO: ORIGIN CHECK
	},
}