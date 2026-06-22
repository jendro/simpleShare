package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type messagePayload struct {
	Room     string `json:"room"`
	Message  string `json:"message"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	FileName string `json:"fileName"`
	FileType string `json:"fileType"`
	FileSize string `json:"fileSize"`
	FileData string `json:"fileData"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

func wsHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room, password := roomFromQuery(r)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade: %v", err)
			return
		}
		if err := hub.ensureRoom(room, password); err != nil {
			log.Printf("room auth failed: %v", err)
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, err.Error()))
			conn.Close()
			return
		}
		hub.addClient(conn, room)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read: %v", err)
				hub.removeClient(conn)
				break
			}

			var payload messagePayload
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Printf("json unmarshal: %v", err)
				continue
			}
			if payload.Message == "" {
				continue
			}
			hub.broadcast(string(msg), room)
		}
	}
}
