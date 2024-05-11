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

	clients = make(map[*websocket.Conn]bool)
)

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)

	clients[conn] = true

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			return
		}

		fmt.Printf("%s send: %s\n", conn.RemoteAddr(), string(msg))
		for client := range clients {
			if client != conn {
				if err := client.WriteMessage(msgType, msg); err != nil {
					delete(clients, client)
					return
				}
			}
		}
	}
}

func Websocketpage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/room.html")
}

func Websocketpage2(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/rooms.html")
}
