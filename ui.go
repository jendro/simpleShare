package main

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
        <button id="openProfileButton" class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 shadow-sm hover:border-slate-900">
          <span id="profileAvatar" class="inline-flex h-9 w-9 items-center justify-center rounded-2xl bg-slate-900 text-lg text-white">👾</span>
          <span id="profileName" class="font-medium">Guest</span>
        </button>
        <button id="openRoomButton" class="inline-flex items-center rounded-full bg-slate-900 text-white px-4 py-2 text-sm font-medium hover:bg-slate-800">Room</button>
      </div>
    </div>
    <div class="p-6 space-y-4">
      <div class="flex flex-col gap-3">
        <textarea id="input" class="min-h-[220px] w-full rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm font-mono text-slate-900 focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200" placeholder="Tempel teks, JSON, atau seret file ke sini lalu tekan Enter..." disabled></textarea>
        <input id="fileInput" type="file" class="hidden" />
        <div id="filePreview" class="hidden items-center justify-between gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700"></div>
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

<div id="profileModal" class="fixed inset-0 z-50 hidden items-center justify-center bg-slate-900/50 px-4 py-6">
  <div class="w-full max-w-md rounded-3xl bg-white p-6 shadow-2xl border border-slate-200">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h2 class="text-lg font-semibold">Ubah Profil</h2>
        <p class="text-sm text-slate-500">Ubah username dan avatar Anda tanpa mengganti room.</p>
      </div>
      <button id="closeProfileButton" class="text-slate-500 hover:text-slate-900">Tutup</button>
    </div>
    <div class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-slate-700">Username</label>
        <input id="profileUsername" class="mt-1 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm text-slate-900 focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200" placeholder="Contoh: dev123" />
      </div>
      <div>
        <label class="block text-sm font-medium text-slate-700">Pilih Avatar</label>
        <div class="mt-2 grid grid-cols-4 gap-2" id="profileAvatarOptions">
          <button type="button" data-avatar="👾" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">👾</button>
          <button type="button" data-avatar="🧑‍💻" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🧑‍💻</button>
          <button type="button" data-avatar="🎮" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🎮</button>
          <button type="button" data-avatar="🤖" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🤖</button>
          <button type="button" data-avatar="🐉" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🐉</button>
          <button type="button" data-avatar="👻" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">👻</button>
          <button type="button" data-avatar="🦾" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🦾</button>
          <button type="button" data-avatar="🧙" class="profileAvatarOption rounded-2xl border border-slate-300 bg-slate-50 p-3 text-2xl transition hover:border-slate-900">🧙</button>
        </div>
      </div>
      <div class="flex justify-end gap-3">
        <button id="cancelProfileButton" class="rounded-2xl border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-100">Batal</button>
        <button id="saveProfileButton" class="rounded-2xl bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800">Simpan</button>
      </div>
    </div>
  </div>
</div>
<script src="/static/app.js"></script>
</body>
</html>`
