<template>
  <div class="session-list">
    <button class="new" :disabled="disabled" @click="$emit('create')">+ 新建会话</button>
    <ul>
      <li
        v-for="session in sessions"
        :key="session.id"
        :class="{ active: session.id === selectedId }"
        @click="$emit('select', session.id)"
      >
        <div class="name">{{ session.name }}</div>
        <div class="meta">{{ formatTime(session.lastActiveAt) }}</div>
        <button class="rename" @click.stop="$emit('rename', session.id)">改名</button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
interface SessionItem {
  id: string;
  name: string;
  createdAt: string;
  lastInputSummary: string;
  lastActiveAt: string;
}

defineProps<{ sessions: SessionItem[]; selectedId: string | null; disabled: boolean }>();

defineEmits<{ (event: "select", sessionId: string): void; (event: "create"): void; (event: "rename", sessionId: string): void }>();

const formatTime = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toLocaleString();
};
</script>

<style scoped>
.session-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.new {
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: transparent;
  color: inherit;
  font-family: "IBM Plex Mono", monospace;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  border-radius: 8px;
  cursor: pointer;
}

.new:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

li {
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 10px;
  background: rgba(9, 10, 12, 0.7);
  cursor: pointer;
  position: relative;
}

li.active {
  border-color: rgba(158, 227, 125, 0.4);
  box-shadow: 0 0 20px rgba(158, 227, 125, 0.15);
}

.name {
  font-family: "IBM Plex Mono", monospace;
  font-size: 13px;
}

.meta {
  margin-top: 6px;
  font-size: 11px;
  color: #8c8f96;
}

.rename {
  position: absolute;
  top: 10px;
  right: 10px;
  background: transparent;
  border: none;
  color: #8c8f96;
  font-size: 11px;
  cursor: pointer;
}
</style>
