package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	connc map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		connc: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWSOrderbook(ws *websocket.Conn) {

	fmt.Println("New connection to orderbook", ws.RemoteAddr())

	for {
		payload := fmt.Sprintf("Orderbook data -> %d\n", time.Now().UnixNano())
		ws.Write([]byte(payload))
		time.Sleep(time.Second * 2)

	}

}

func (s *Server) wsHandler(ws *websocket.Conn) {
	fmt.Println("New connection", ws.RemoteAddr())

	s.connc[ws] = true

	s.readLoop(ws)

}

func (s *Server) readLoop(ws *websocket.Conn) {
	buff := make([]byte, 1024)

	for {
		n, err := ws.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading:", err)
			continue
		}

		msg := buff[:n]
		s.broadcast(msg)
	}

}

func (s *Server) broadcast(b []byte) {
	for ws := range s.connc {
		go func(ws *websocket.Conn) {
			var err error
			if _, err = ws.Write(b); err != nil {
				fmt.Println("Error writing:", err)
			}
		}(ws)
	}
}

func main() {
	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.wsHandler))
	http.Handle("/wsorderbook", websocket.Handler(server.handleWSOrderbook))
	http.ListenAndServe(":3000", nil)
}
