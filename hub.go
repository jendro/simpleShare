package main

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	rooms     map[string]map[*websocket.Conn]bool
	passwords map[string]string
	connRoom  map[*websocket.Conn]string
	mu        sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		rooms:     make(map[string]map[*websocket.Conn]bool),
		passwords: make(map[string]string),
		connRoom:  make(map[*websocket.Conn]string),
	}
}

func (h *Hub) broadcast(message, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.rooms[room] {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
			delete(h.rooms[room], conn)
			delete(h.connRoom, conn)
		}
	}
}

func (h *Hub) addClient(conn *websocket.Conn, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if oldRoom, ok := h.connRoom[conn]; ok && oldRoom != room {
		if conns, ok := h.rooms[oldRoom]; ok {
			delete(conns, conn)
			if len(conns) == 0 {
				delete(h.rooms, oldRoom)
			}
		}
	}

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*websocket.Conn]bool)
	}
	h.rooms[room][conn] = true
	h.connRoom[conn] = room
}

func (h *Hub) removeClient(conn *websocket.Conn) {
	h.mu.Lock()
	room, ok := h.connRoom[conn]
	if ok {
		if conns, ok := h.rooms[room]; ok {
			delete(conns, conn)
			if len(conns) == 0 {
				delete(h.rooms, room)
			}
		}
		delete(h.connRoom, conn)
	}
	h.mu.Unlock()
	conn.Close()
}

func (h *Hub) ensureRoom(room, password string) error {
	if room == "" {
		room = "global"
	}
	if saved, ok := h.passwords[room]; ok {
		if saved != password {
			return errors.New("password room salah")
		}
		return nil
	}
	if password != "" {
		h.passwords[room] = password
	}
	return nil
}

func roomFromQuery(r *http.Request) (string, string) {
	room := r.URL.Query().Get("room")
	password := r.URL.Query().Get("password")
	if room == "" {
		room = "global"
	}
	return room, password
}
