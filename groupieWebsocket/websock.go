package groupieWebsocket

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	clients = make([]*websocket.Conn, 0)
)

func Websocket(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)

	clients = append(clients, conn)

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		fmt.Printf("%s send: %s\n", conn.RemoteAddr(), string(msg))
		for _, client := range clients {
			if err := client.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}
}

func Websocketpage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/room.html")
}
