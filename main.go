package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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
			log.Printf("broadcast write error: %v", err)
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
		room = "public"
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

func main() {
	flag.Parse()
	hub := newHub()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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

			var payload struct {
				Room    string `json:"room"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Printf("json unmarshal: %v", err)
				continue
			}
			if payload.Message == "" {
				continue
			}
			hub.broadcast(payload.Message, payload.Room)
		}
	})

	log.Printf("starting server at %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}

func roomFromQuery(r *http.Request) (string, string) {
	room := r.URL.Query().Get("room")
	password := r.URL.Query().Get("password")
	if room == "" {
		room = "public"
	}
	return room, password
}

const indexHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Shared Text & JSON Share</title>
<style>
body { font-family: system-ui, sans-serif; margin: 0; padding: 0; background: #f7f7f7; }
.container { max-width: 760px; margin: 2rem auto; padding: 1rem; background: #fff; border-radius: 12px; box-shadow: 0 10px 30px rgba(0,0,0,0.08); }
h1 { margin-top: 0; }
textarea { width: 100%; min-height: 150px; padding: 12px; border: 1px solid #ccc; border-radius: 8px; font-family: monospace; font-size: 14px; resize: vertical; }
#messages { margin-top: 1rem; display: grid; gap: 10px; }
.message { padding: 1rem; border: 1px solid #e2e8f0; border-radius: 10px; background: #f9fafb; display: flex; justify-content: space-between; gap: 0.75rem; align-items: flex-start; }
.message pre { margin: 0; white-space: pre-wrap; word-break: break-word; flex: 1; }
button { background: #2563eb; color: white; border: none; border-radius: 8px; padding: 0.6rem 0.9rem; cursor: pointer; }
button:disabled { opacity: 0.5; cursor: default; }
.status { margin-top: 0.75rem; color: #555; }
</style>
</head>
<body>
<div class="container">
<h1>Shared Text / JSON</h1>
<p>Tempel teks, JSON, atau konten apa saja ke textarea, lalu tekan Enter untuk dibagikan ke semua perangkat di room ini.</p>
<div class="room-controls">
  <input id="room" placeholder="Nama room (kosong = public)" />
  <input id="password" type="password" placeholder="Password room (opsional)" />
  <button id="joinButton" type="button">Join Room</button>
</div>
<textarea id="input" placeholder="Tempel teks, JSON, atau apa saja lalu tekan Enter..." disabled></textarea>
<div class="status" id="status">Masukkan room dan tekan Join Room.</div>
<div id="messages"></div>
</div>
<script>
const roomInput = document.getElementById('room');
const passwordInput = document.getElementById('password');
const joinButton = document.getElementById('joinButton');
const input = document.getElementById('input');
const status = document.getElementById('status');
const messages = document.getElementById('messages');
let socket;
let currentRoom = 'public';
let currentPassword = '';

function setStatus(text) {
  status.textContent = text;
}

async function copyToClipboard(text) {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    return navigator.clipboard.writeText(text);
  }

  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed';
  textarea.style.top = '-9999px';
  textarea.style.left = '-9999px';
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();

  const successful = document.execCommand('copy');
  document.body.removeChild(textarea);

  if (!successful) {
    throw new Error('copy command failed');
  }
}

function parseQuery() {
  const url = new URL(window.location.href);
  return {
    room: url.searchParams.get('room') || 'public',
    password: url.searchParams.get('password') || '',
  };
}

function connect(room, password) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.close();
  }

  currentRoom = room || 'public';
  currentPassword = password || '';
  const params = '?room=' + encodeURIComponent(currentRoom) + '&password=' + encodeURIComponent(currentPassword);
  socket = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws' + params);

  socket.addEventListener('open', function () {
    setStatus('Terhubung ke room \' + currentRoom + '\'. Ketik lalu Enter untuk membagikan.');
    input.disabled = false;
    input.focus();
  });
  socket.addEventListener('close', () => {
    setStatus('Terputus. Tekan Join Room untuk mencoba lagi.');
    input.disabled = true;
  });
  socket.addEventListener('error', () => {
    setStatus('Gagal terhubung. Periksa nama room dan password.');
    input.disabled = true;
  });
  socket.addEventListener('message', event => addMessage(event.data));
}

function addMessage(text) {
  const wrapper = document.createElement('div');
  wrapper.className = 'message';

  const pre = document.createElement('pre');
  pre.textContent = text;

  const copyButton = document.createElement('button');
  copyButton.textContent = 'Copy';
  copyButton.type = 'button';
  copyButton.addEventListener('click', async () => {
    try {
      await copyToClipboard(text);
      copyButton.textContent = 'Copied';
      setTimeout(() => { copyButton.textContent = 'Copy'; }, 1200);
    } catch (err) {
      console.error(err);
      copyButton.textContent = 'Gagal';
      setTimeout(() => { copyButton.textContent = 'Copy'; }, 1500);
    }
  });

  wrapper.appendChild(pre);
  wrapper.appendChild(copyButton);
  messages.prepend(wrapper);
}

joinButton.addEventListener('click', () => {
  const room = roomInput.value.trim() || 'public';
  const password = passwordInput.value;
  connect(room, password);
});

input.addEventListener('keydown', event => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault();
    const text = input.value.trim();
    if (!text || !socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }
    socket.send(JSON.stringify({ room: currentRoom, message: text }));
    input.value = '';
  }
});

const initial = parseQuery();
if (initial.room !== 'public' || initial.password !== '') {
  roomInput.value = initial.room === 'public' ? '' : initial.room;
  passwordInput.value = initial.password;
  connect(initial.room, initial.password);
}
</script>
</body>
</html>`
