<template>
  <div class="terminal" ref="terminalRef"></div>
</template>

<script setup lang="ts">
import "xterm/css/xterm.css";

import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";

import { loadTerminalCache, saveTerminalCache } from "../utils/terminalCache";

const props = defineProps<{ sessionId: string | null; token: string }>();
const emit = defineEmits<{ (event: "connected"): void; (event: "disconnected"): void }>();

const terminalRef = ref<HTMLDivElement | null>(null);
let terminal: Terminal | null = null;
let fitAddon: FitAddon | null = null;
let socket: WebSocket | null = null;
let outputBuffer = "";

const connectSocket = (sessionId: string) => {
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = `${protocol}://${window.location.host}/ws/terminal/${sessionId}`;
  socket = new WebSocket(wsUrl);

  socket.onopen = () => {
    emit("connected");
  };

  socket.onclose = () => {
    emit("disconnected");
  };

  socket.onmessage = (event) => {
    if (!terminal) {
      return;
    }
    terminal.write(event.data);
    outputBuffer += event.data;
    // Limit cached output to avoid unbounded growth.
    if (outputBuffer.length > 20000) {
      outputBuffer = outputBuffer.slice(-20000);
    }
    saveTerminalCache(sessionId, outputBuffer);
  };
};

const initTerminal = (sessionId: string) => {
  if (!terminalRef.value) {
    return;
  }

  terminal = new Terminal({
    fontFamily: "IBM Plex Mono, monospace",
    fontSize: 13,
    theme: {
      background: "#0a0b0c",
      foreground: "#e8e8e1",
      cursor: "#9ee37d",
    },
  });

  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.open(terminalRef.value);
  fitAddon.fit();

  outputBuffer = loadTerminalCache(sessionId);
  if (outputBuffer) {
    terminal.write(outputBuffer);
  }

  terminal.onData((data) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(data);
    }
  });
};

const disposeTerminal = () => {
  socket?.close();
  socket = null;
  terminal?.dispose();
  terminal = null;
  fitAddon = null;
};

watch(
  () => props.sessionId,
  (sessionId) => {
    disposeTerminal();
    if (sessionId) {
      initTerminal(sessionId);
      connectSocket(sessionId);
    }
  },
  { immediate: true },
);

onMounted(() => {
  const resize = () => fitAddon?.fit();
  window.addEventListener("resize", resize);
  onBeforeUnmount(() => window.removeEventListener("resize", resize));
});

onBeforeUnmount(() => {
  disposeTerminal();
});
</script>

<style scoped>
.terminal {
  width: 100%;
  height: 100%;
  min-height: 70vh;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  background: #0a0b0c;
  padding: 12px;
}
</style>
