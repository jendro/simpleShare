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
				Room     string `json:"room"`
				Message  string `json:"message"`
				Username string `json:"username"`
				Avatar   string `json:"avatar"`
			}
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Printf("json unmarshal: %v", err)
				continue
			}
			if payload.Message == "" {
				continue
			}
			hub.broadcast(string(msg), room)
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
		room = "global"
	}
	return room, password
}

const indexHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Shared Text & JSON Share</title>
<script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-50 text-slate-800">
<div class="min-h-screen flex items-center justify-center p-4">
  <div class="w-full max-w-4xl bg-white shadow-xl rounded-3xl border border-slate-200 overflow-hidden">
    <div class="flex items-center justify-between px-6 py-5 border-b border-slate-200">
      <div>
        <h1 class="text-2xl font-semibold">Shared Text / JSON</h1>
        <p class="text-sm text-slate-500 mt-1">Secara otomatis terhubung ke room default <strong>global</strong>. Klik tombol Room untuk membuat atau bergabung ke room privat.</p>
      </div>
      <div class="flex items-center gap-3">
        <div id="profileDisplay" class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 shadow-sm">
          <span id="profileAvatar" class="inline-flex h-9 w-9 items-center justify-center rounded-2xl bg-slate-900 text-lg text-white">👾</span>
          <span id="profileName" class="font-medium">Guest</span>
        </div>
        <button id="openRoomButton" class="inline-flex items-center rounded-full bg-slate-900 text-white px-4 py-2 text-sm font-medium hover:bg-slate-800">Room</button>
      </div>
    </div>
    <div class="p-6 space-y-4">
      <div class="flex flex-col gap-3">
        <textarea id="input" class="min-h-[220px] w-full rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm font-mono text-slate-900 focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200" placeholder="Tempel teks, JSON, atau apa saja lalu tekan Enter..." disabled></textarea>
        <div class="text-sm text-slate-500" id="status">Memuat room global...</div>
      </div>
      <div id="messages" class="grid gap-4"></div>
    </div>
  </div>
</div>

<div id="roomModal" class="fixed inset-0 z-50 hidden items-center justify-center bg-slate-900/50 px-4 py-6">
  <div class="w-full max-w-md rounded-3xl bg-white p-6 shadow-2xl border border-slate-200">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h2 class="text-lg font-semibold">Masuk Room</h2>
        <p class="text-sm text-slate-500">Buat room privat dengan password atau bergabung ke room yang sudah ada.</p>
      </div>
      <button id="closeRoomButton" class="text-slate-500 hover:text-slate-900">Tutup</button>
    </div>
    <div class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-slate-700">Nama Room</label>
        <input id="room" class="mt-1 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm text-slate-900 focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200" placeholder="Contoh: project-a" />
      </div>
      <div>
        <label class="block text-sm font-medium text-slate-700">Username</label>
        <input id="username" class="mt-1 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm text-slate-900 focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200" placeholder="Contoh: dev123" />
      </div>
      <div>
        <label class="block text-sm font-medium text-slate-700">Pilih Avatar</label>
        <div class="mt-2 grid grid-cols-4 gap-2" id="avatarOptions">
          <button type="button" data-avatar="👾" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">👾</button>
          <button type="button" data-avatar="🧑‍💻" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🧑‍💻</button>
          <button type="button" data-avatar="🎮" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🎮</button>
          <button type="button" data-avatar="🤖" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🤖</button>
          <button type="button" data-avatar="🐉" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🐉</button>
          <button type="button" data-avatar="👻" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">👻</button>
          <button type="button" data-avatar="🦾" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🦾</button>
          <button type="button" data-avatar="🧙" class="avatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🧙</button>
        </div>
      </div>
      <div>
        <label class="block text-sm font-medium text-slate-700">Password (opsional)</label>
        <input id="password" type="password" class="mt-1 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm text-slate-900 focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200" placeholder="Password room" />
      </div>
      <div class="flex justify-end gap-3">
        <button id="cancelRoomButton" class="rounded-2xl border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-100">Batal</button>
        <button id="joinButton" class="rounded-2xl bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800">Join Room</button>
      </div>
    </div>
  </div>
</div>
<script>
const openRoomButton = document.getElementById('openRoomButton');
const roomModal = document.getElementById('roomModal');
const closeRoomButton = document.getElementById('closeRoomButton');
const cancelRoomButton = document.getElementById('cancelRoomButton');
const roomInput = document.getElementById('room');
const usernameInput = document.getElementById('username');
const passwordInput = document.getElementById('password');
const avatarOptions = document.querySelectorAll('.avatarOption');
const joinButton = document.getElementById('joinButton');
const input = document.getElementById('input');
const status = document.getElementById('status');
const messages = document.getElementById('messages');
const profileAvatar = document.getElementById('profileAvatar');
const profileName = document.getElementById('profileName');
let socket;
let currentRoom = 'global';
let currentPassword = '';
let currentUsername = 'Guest';
let currentAvatar = '👾';
const avatarChoices = ['👾', '🧑‍💻', '🎮', '🤖', '🐉', '👻', '🦾', '🧙'];

function setStatus(text) {
  status.textContent = text;
}

function updateProfileDisplay() {
  profileAvatar.textContent = currentAvatar;
  profileName.textContent = currentUsername;
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
    room: url.searchParams.get('room') || 'global',
    password: url.searchParams.get('password') || '',
  };
}

function connect(room, password) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.close();
  }

  currentRoom = room || 'global';
  currentPassword = password || '';
  const params = '?room=' + encodeURIComponent(currentRoom) + '&password=' + encodeURIComponent(currentPassword);
  socket = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws' + params);

  socket.addEventListener('open', function () {
    setStatus('Terhubung ke room ' + currentRoom + '. Username: ' + currentUsername);
    updateProfileDisplay();
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
  let payload = { username: 'Guest', avatar: '👾', message: text };
  try {
    payload = JSON.parse(text);
  } catch (err) {
    // fallback to plain text
  }

  const wrapper = document.createElement('div');
  wrapper.className = 'rounded-3xl border border-slate-200 bg-slate-50 p-4 shadow-sm';

  const header = document.createElement('div');
  header.className = 'flex items-center justify-between gap-3';

  const userInfo = document.createElement('div');
  userInfo.className = 'flex items-center gap-3';

  const avatar = document.createElement('div');
  avatar.className = 'flex h-10 w-10 items-center justify-center rounded-2xl bg-slate-900 text-lg text-white';
  avatar.textContent = payload.avatar || '👾';

  const name = document.createElement('div');
  name.className = 'flex flex-col';

  const usernameEl = document.createElement('span');
  usernameEl.className = 'text-sm font-semibold text-slate-900';
  usernameEl.textContent = payload.username || 'Guest';

  const roomEl = document.createElement('span');
  roomEl.className = 'text-xs text-slate-500';
  roomEl.textContent = payload.room ? payload.room : (currentRoom === 'global' ? 'Global' : currentRoom);

  name.appendChild(usernameEl);
  name.appendChild(roomEl);
  userInfo.appendChild(avatar);
  userInfo.appendChild(name);

  const copyButton = document.createElement('button');
  copyButton.innerHTML = '📋';
  copyButton.type = 'button';
  copyButton.className = 'inline-flex h-10 w-10 items-center justify-center rounded-full border border-slate-200 bg-white text-lg text-slate-700 transition hover:bg-slate-100';
  copyButton.addEventListener('click', async () => {
    try {
      await copyToClipboard(payload.message || text);
      copyButton.textContent = '✅';
      setTimeout(() => { copyButton.innerHTML = '📋'; }, 1200);
    } catch (err) {
      console.error(err);
      copyButton.textContent = '❌';
      setTimeout(() => { copyButton.innerHTML = '📋'; }, 1500);
    }
  });

  header.appendChild(userInfo);
  header.appendChild(copyButton);

  const pre = document.createElement('pre');
  pre.textContent = payload.message || text;
  pre.className = 'mt-3 text-sm text-slate-700 whitespace-pre-wrap break-words';

  wrapper.appendChild(header);
  wrapper.appendChild(pre);
  messages.prepend(wrapper);
}

openRoomButton.addEventListener('click', () => {
  roomModal.classList.remove('hidden');
});

closeRoomButton.addEventListener('click', () => {
  roomModal.classList.add('hidden');
});

cancelRoomButton.addEventListener('click', () => {
  roomModal.classList.add('hidden');
});

avatarOptions.forEach((button, index) => {
  button.addEventListener('click', () => {
    avatarOptions.forEach(el => el.classList.remove('border-slate-900', 'ring-2', 'ring-slate-900'));
    button.classList.add('border-slate-900', 'ring-2', 'ring-slate-900');
    currentAvatar = button.dataset.avatar;
    updateProfileDisplay();
  });
});

if (avatarOptions.length > 0) {
  avatarOptions[0].classList.add('border-slate-900', 'ring-2', 'ring-slate-900');
  currentAvatar = avatarOptions[0].dataset.avatar;
}

usernameInput.addEventListener('input', () => {
  currentUsername = usernameInput.value.trim() || 'Guest';
  updateProfileDisplay();
});

joinButton.addEventListener('click', () => {
  const room = roomInput.value.trim() || 'global';
  currentUsername = usernameInput.value.trim() || 'Guest';
  const password = passwordInput.value;
  if (room === currentRoom && socket && socket.readyState === WebSocket.OPEN) {
    updateProfileDisplay();
    setStatus('Username/avatar diperbarui di room ' + currentRoom);
  } else {
    connect(room, password);
  }
  roomModal.classList.add('hidden');
});

input.addEventListener('keydown', event => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault();
    const text = input.value.trim();
    if (!text || !socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }
    socket.send(JSON.stringify({ room: currentRoom, message: text, username: currentUsername, avatar: currentAvatar }));
    input.value = '';
  }
});

const initial = parseQuery();
roomInput.value = initial.room === 'global' ? '' : initial.room;
passwordInput.value = initial.password;
usernameInput.value = currentUsername;
updateProfileDisplay();
connect(initial.room, initial.password);
</script>
</body>
</html>`
