<template>
  <div class="app">
    <header class="topbar">
      <div class="brand">
        <span class="brand-dot"></span>
        Anywhere Code
      </div>
      <div class="status" :class="{ online: isConnected }">
        {{ isConnected ? "已连接" : "未连接" }}
      </div>
    </header>

    <div class="layout">
      <aside class="sidebar">
        <div class="panel-title">会话</div>
        <SessionList
          :sessions="sessions"
          :selected-id="selectedSessionId"
          :disabled="!isAuthed"
          @select="selectSession"
          @create="createSession"
          @rename="renameSession"
        />
      </aside>

      <main class="main">
        <div v-if="!isAuthed" class="login">
          <h2>登录</h2>
          <form @submit.prevent="login">
            <label>
              用户名
              <input v-model="loginForm.username" autocomplete="username" />
            </label>
            <label>
              密码
              <input v-model="loginForm.password" type="password" autocomplete="current-password" />
            </label>
            <button type="submit">进入终端</button>
            <p v-if="loginError" class="error">{{ loginError }}</p>
          </form>
        </div>

        <div v-else class="terminal-wrap">
          <TerminalView
            :session-id="selectedSessionId"
            :token="token"
            @connected="isConnected = true"
            @disconnected="isConnected = false"
          />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";

import SessionList from "./components/SessionList.vue";
import TerminalView from "./components/TerminalView.vue";

interface SessionItem {
  id: string;
  name: string;
  createdAt: string;
  lastInputSummary: string;
  lastActiveAt: string;
}

const sessions = ref<SessionItem[]>([]);
const selectedSessionId = ref<string | null>(null);
const token = ref<string>(localStorage.getItem("token") || "");
const isAuthed = ref(Boolean(token.value));
const isConnected = ref(false);
const loginError = ref("");

const loginForm = ref({
  username: "",
  password: "",
});

const authHeaders = () => ({
  Authorization: `Bearer ${token.value}`,
});

const loadSessions = async () => {
  const res = await fetch("/api/sessions", { headers: authHeaders() });
  if (!res.ok) {
    return;
  }
  const data = (await res.json()) as SessionItem[];
  sessions.value = data;
  if (!selectedSessionId.value && data.length > 0) {
    selectedSessionId.value = data[0].id;
  }
};

const login = async () => {
  loginError.value = "";
  const res = await fetch("/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(loginForm.value),
  });

  if (!res.ok) {
    loginError.value = "账号或密码错误";
    return;
  }

  const data = await res.json();
  token.value = data.token;
  localStorage.setItem("token", token.value);
  isAuthed.value = true;
  await loadSessions();
};

const createSession = async () => {
  const res = await fetch("/api/sessions", {
    method: "POST",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  if (!res.ok) {
    return;
  }
  const session = (await res.json()) as SessionItem;
  sessions.value = [session, ...sessions.value];
  selectedSessionId.value = session.id;
};

const renameSession = async (sessionId: string) => {
  const target = sessions.value.find((item) => item.id === sessionId);
  if (!target) {
    return;
  }
  const name = window.prompt("新的会话名", target.name);
  if (!name) {
    return;
  }
  const res = await fetch(`/api/sessions/${sessionId}/rename`, {
    method: "POST",
    headers: { ...authHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) {
    return;
  }
  const updated = (await res.json()) as SessionItem;
  sessions.value = sessions.value.map((item) => (item.id === sessionId ? updated : item));
};

const selectSession = (sessionId: string) => {
  selectedSessionId.value = sessionId;
};

onMounted(async () => {
  if (isAuthed.value) {
    await loadSessions();
  }
});
</script>

<style>
@import url("https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@300;400;500;700&family=Manrope:wght@300;500;700&display=swap");

:root {
  color-scheme: dark;
  --bg: #0a0b0c;
  --panel: #111316;
  --panel-2: #0d0f12;
  --text: #e8e8e1;
  --muted: #8c8f96;
  --accent: #9ee37d;
  --danger: #ff6b6b;
  --border: rgba(255, 255, 255, 0.08);
  --glow: rgba(158, 227, 125, 0.4);
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: "Manrope", sans-serif;
  background: radial-gradient(circle at top, #0e1115 0%, #070809 45%, #040405 100%);
  color: var(--text);
}

.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 28px;
  border-bottom: 1px solid var(--border);
  background: rgba(6, 6, 7, 0.9);
  backdrop-filter: blur(8px);
}

.brand {
  font-family: "IBM Plex Mono", monospace;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-dot {
  width: 10px;
  height: 10px;
  background: var(--accent);
  border-radius: 50%;
  box-shadow: 0 0 12px var(--glow);
}

.status {
  font-size: 12px;
  color: var(--muted);
}

.status.online {
  color: var(--accent);
}

.layout {
  flex: 1;
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 0;
}

.sidebar {
  background: var(--panel);
  padding: 24px 18px;
  border-right: 1px solid var(--border);
}

.panel-title {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.16em;
  color: var(--muted);
  margin-bottom: 16px;
}

.main {
  background: var(--panel-2);
  display: flex;
  align-items: stretch;
  justify-content: stretch;
}

.login {
  margin: auto;
  background: rgba(12, 14, 17, 0.9);
  padding: 32px;
  border: 1px solid var(--border);
  border-radius: 16px;
  width: min(360px, 90%);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.4);
}

.login h2 {
  margin-top: 0;
  font-weight: 600;
}

.login label {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 12px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.login input {
  background: transparent;
  border: 1px solid var(--border);
  padding: 10px 12px;
  color: var(--text);
  font-family: "IBM Plex Mono", monospace;
  border-radius: 8px;
}

.login button {
  width: 100%;
  padding: 12px;
  background: var(--accent);
  border: none;
  color: #0b0d0f;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  border-radius: 8px;
  cursor: pointer;
}

.error {
  color: var(--danger);
  margin-top: 12px;
}

.terminal-wrap {
  flex: 1;
  padding: 16px;
}

@media (max-width: 900px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .sidebar {
    border-right: none;
    border-bottom: 1px solid var(--border);
  }
}
</style>
