const openRoomButton = document.getElementById('openRoomButton');
const roomModal = document.getElementById('roomModal');
const closeRoomButton = document.getElementById('closeRoomButton');
const cancelRoomButton = document.getElementById('cancelRoomButton');
const roomInput = document.getElementById('room');
const passwordInput = document.getElementById('password');
const joinButton = document.getElementById('joinButton');
const openProfileButton = document.getElementById('openProfileButton');
const profileModal = document.getElementById('profileModal');
const closeProfileButton = document.getElementById('closeProfileButton');
const cancelProfileButton = document.getElementById('cancelProfileButton');
const saveProfileButton = document.getElementById('saveProfileButton');
const profileUsernameInput = document.getElementById('profileUsername');
const profileAvatarOptions = document.querySelectorAll('.profileAvatarOption');
const input = document.getElementById('input');
const fileInput = document.getElementById('fileInput');
const filePreview = document.getElementById('filePreview');
const status = document.getElementById('status');
const messages = document.getElementById('messages');
const profileAvatar = document.getElementById('profileAvatar');
const profileName = document.getElementById('profileName');
let socket;
let currentRoom = 'global';
let currentPassword = '';
let currentUsername = 'Guest';
let currentAvatar = '👾';

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

let pendingFile = null;

function showFilePreview(file) {
  filePreview.classList.remove('hidden');
  filePreview.innerHTML = '';

  const fileText = document.createElement('div');
  fileText.className = 'flex-1 overflow-hidden text-ellipsis whitespace-nowrap';
  fileText.textContent = '📎 ' + file.name + ' (' + Math.round(file.size / 1024) + ' KB)';

  const removeButton = document.createElement('button');
  removeButton.type = 'button';
  removeButton.className = 'rounded-full border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 hover:bg-slate-100';
  removeButton.textContent = 'Hapus';
  removeButton.addEventListener('click', clearPendingFile);

  filePreview.appendChild(fileText);
  filePreview.appendChild(removeButton);
}

function clearPendingFile() {
  pendingFile = null;
  filePreview.classList.add('hidden');
  filePreview.innerHTML = '';
  fileInput.value = '';
}

function setPendingFile(file) {
  pendingFile = file;
  showFilePreview(file);
}

function handleFileSelection(file) {
  if (!file) {
    return;
  }
  setPendingFile(file);
}

function handleDrop(event) {
  event.preventDefault();
  const file = event.dataTransfer.files && event.dataTransfer.files[0];
  if (file) {
    handleFileSelection(file);
    setStatus('File siap dikirim: ' + file.name + '. Tekan Enter untuk kirim.');
  }
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

  const content = document.createElement('div');
  content.className = 'mt-3 relative overflow-hidden rounded-3xl border border-slate-200 bg-white p-4 max-w-full';
  content.style.maxHeight = '200px';

  const pre = document.createElement('pre');
  pre.textContent = payload.message || text;
  pre.className = 'text-sm text-slate-700 whitespace-pre-wrap break-words break-all';
  pre.style.margin = 0;
  pre.style.width = '100%';
  pre.style.overflowWrap = 'anywhere';
  pre.style.wordBreak = 'break-word';

  const fade = document.createElement('div');
  fade.className = 'pointer-events-none absolute inset-x-0 bottom-0 h-16 bg-gradient-to-t from-white to-transparent';

  content.appendChild(pre);
  content.appendChild(fade);

  wrapper.appendChild(header);
  wrapper.appendChild(content);

  if (payload.fileName && payload.fileData) {
    const fileWrapper = document.createElement('div');
    fileWrapper.className = 'mt-3 rounded-2xl border border-slate-200 bg-slate-50 p-4';

    const fileTitle = document.createElement('div');
    fileTitle.className = 'mb-2 text-sm font-semibold text-slate-900';
    fileTitle.textContent = 'File dibagikan:';

    const fileLink = document.createElement('a');
    fileLink.href = payload.fileData;
    fileLink.download = payload.fileName;
    fileLink.target = '_blank';
    fileLink.rel = 'noreferrer noopener';
    fileLink.className = 'text-sm text-slate-700 underline hover:text-slate-900';
    fileLink.textContent = payload.fileName + (payload.fileSize ? ' (' + payload.fileSize + ')' : '');

    fileWrapper.appendChild(fileTitle);
    fileWrapper.appendChild(fileLink);
    wrapper.appendChild(fileWrapper);
  }

  let expandButton;
  messages.prepend(wrapper);

  requestAnimationFrame(() => {
    if (pre.scrollHeight <= 200) {
      fade.remove();
      return;
    }

    expandButton = document.createElement('button');
    expandButton.type = 'button';
    expandButton.textContent = 'Selengkapnya';
    expandButton.className = 'mt-3 text-sm font-medium text-slate-700 hover:text-slate-900';

    let expanded = false;
    expandButton.addEventListener('click', () => {
      expanded = !expanded;
      if (expanded) {
        content.style.maxHeight = 'none';
        content.classList.remove('overflow-hidden');
        fade.style.display = 'none';
        expandButton.textContent = 'Sembunyikan';
      } else {
        content.style.maxHeight = '200px';
        content.classList.add('overflow-hidden');
        fade.style.display = '';
        expandButton.textContent = 'Selengkapnya';
      }
    });

    wrapper.appendChild(expandButton);
  });
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

openProfileButton.addEventListener('click', () => {
  profileModal.classList.remove('hidden');
  profileUsernameInput.value = currentUsername;
  profileAvatarOptions.forEach(el => el.classList.remove('border-slate-900', 'ring-2', 'ring-slate-900'));
  profileAvatarOptions.forEach(el => {
    if (el.dataset.avatar === currentAvatar) {
      el.classList.add('border-slate-900', 'ring-2', 'ring-slate-900');
    }
  });
});

closeProfileButton.addEventListener('click', () => {
  profileModal.classList.add('hidden');
});

cancelProfileButton.addEventListener('click', () => {
  profileModal.classList.add('hidden');
});

profileAvatarOptions.forEach(button => {
  button.addEventListener('click', () => {
    profileAvatarOptions.forEach(el => el.classList.remove('border-slate-900', 'ring-2', 'ring-slate-900'));
    button.classList.add('border-slate-900', 'ring-2', 'ring-slate-900');
    currentAvatar = button.dataset.avatar;
  });
});

profileUsernameInput.addEventListener('input', () => {
  currentUsername = profileUsernameInput.value.trim() || 'Guest';
  updateProfileDisplay();
});

saveProfileButton.addEventListener('click', () => {
  currentUsername = profileUsernameInput.value.trim() || 'Guest';
  updateProfileDisplay();
  profileModal.classList.add('hidden');
});

joinButton.addEventListener('click', () => {
  const room = roomInput.value.trim() || 'global';
  const password = passwordInput.value;
  connect(room, password);
  roomModal.classList.add('hidden');
});

input.addEventListener('keydown', event => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault();
    const text = input.value.trim();
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }

    if (pendingFile) {
      const reader = new FileReader();
      reader.onload = () => {
        socket.send(JSON.stringify({
          room: currentRoom,
          message: text || (currentUsername + ' berbagi file'),
          username: currentUsername,
          avatar: currentAvatar,
          fileName: pendingFile.name,
          fileType: pendingFile.type || 'application/octet-stream',
          fileSize: Math.round(pendingFile.size / 1024) + ' KB',
          fileData: reader.result,
        }));
        setStatus('File dikirim: ' + pendingFile.name);
        clearPendingFile();
        input.value = '';
      };
      reader.readAsDataURL(pendingFile);
      return;
    }

    if (!text) {
      return;
    }
    socket.send(JSON.stringify({ room: currentRoom, message: text, username: currentUsername, avatar: currentAvatar }));
    input.value = '';
  }
});

fileInput.addEventListener('change', event => handleFileSelection(event.target.files && event.target.files[0]));
input.addEventListener('dragover', event => {
  event.preventDefault();
  input.classList.add('border-slate-900');
});
input.addEventListener('dragleave', () => {
  input.classList.remove('border-slate-900');
});
input.addEventListener('drop', event => {
  input.classList.remove('border-slate-900');
  handleDrop(event);
});

const initial = parseQuery();
roomInput.value = initial.room === 'global' ? '' : initial.room;
passwordInput.value = initial.password;
profileUsernameInput.value = currentUsername;
updateProfileDisplay();
connect(initial.room, initial.password);
