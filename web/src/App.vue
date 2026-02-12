<template>
  <div class="page">
      <div class="glow"></div>
      <header class="header">
        <div class="brand">
          <span class="brand__title">ANYWHERE CODE</span>
          <span class="brand__subtitle">PTY session pod · WSL runtime</span>
        </div>
        <div class="header__right">
          <n-button class="log-trigger" size="small" @click="showLogDrawer = true">日志</n-button>
          <n-button class="drawer-trigger" size="small" @click="showRightDrawer = true">工作区</n-button>
          <div class="status">
            <span class="status__dot" :class="headerStatusClass"></span>
            <span class="status__text">{{ headerStatus }}</span>
          </div>
        </div>
      </header>

      <div class="workspace" :class="{ 'workspace--collapsed': isSidebarCollapsed }">
        <aside
          class="sidebar"
          :class="{ 'sidebar--collapsed': isSidebarCollapsed }"
          @dragover.prevent
          @drop="handleDropToList"
        >
          <div class="sidebar__header">
            <div>
              <div class="sidebar__title">会话列表</div>
              <div class="sidebar__subtitle">{{ sessionList.length }} 个会话</div>
            </div>
            <div class="sidebar__actions">
              <button class="action" @click="handleNewSession" :disabled="isBusy">新建</button>
              <button
                class="action action--ghost"
                @click="fetchSessions"
                :disabled="isBusy"
                v-show="!isSidebarCollapsed"
              >
                刷新
              </button>
            </div>
          </div>
          <div class="sidebar-loading" v-if="isBusy || isConnecting">
            <span class="spinner spinner--small"></span>
            <span>{{ isBusy ? "正在创建会话..." : "正在连接会话..." }}</span>
          </div>
          <div class="session-list" v-show="!isSidebarCollapsed">
            <button
              v-for="session in sessionList"
              :key="session.id"
              class="session-item"
              :class="{ 'session-item--active': isSessionInPane(session.id) }"
              @click="handleConnectToActivePane(session.id)"
              @contextmenu.prevent="openSessionMenu($event, session)"
              draggable="true"
              @dragstart="handleDragStart($event, session.id)"
            >
              <div class="session-item__id">{{ getListTitle(session) }}</div>
              <div class="session-item__time">活跃：{{ formatTime(session.last_active) }}</div>
            </button>
          </div>
          <div
            v-if="menu.visible"
            class="context-menu"
            :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
            @click.stop
          >
            <button class="context-item" @click="renameSession(menu.session)">重命名</button>
            <button class="context-item context-item--danger" @click="closeSession(menu.session)">关闭会话</button>
          </div>
        </aside>

        <div class="sidebar-toggle-rail">
          <button class="sidebar-toggle" @click="toggleSidebar" aria-label="收起或展开列表">
            <span class="sidebar-toggle__icon">{{ isSidebarCollapsed ? "›" : "‹" }}</span>
          </button>
        </div>

        <section class="main">
          <section
            class="terminal-grid"
            :class="gridClass"
            @dragover.prevent="handleGridDragOver"
            @dragleave="handleGridDragLeave"
            @drop="handleDropToGrid"
          >
            <div
              v-for="pane in visiblePanes"
              :key="pane.slot"
              class="terminal-shell"
              :class="{
                'terminal-shell--active': pane.slot === activePane,
                'terminal-shell--drop': pane.slot === hoverPane
              }"
              @click="activePane = pane.slot"
              @dragover.prevent
              @drop="handleDropToPane($event, pane.slot)"
            >
              <div class="terminal-toolbar">
                <div class="terminal-actions">
                  <button
                    class="terminal-btn terminal-btn--icon terminal-btn--mac terminal-btn--close"
                    @click="handleClosePane(pane)"
                    :disabled="isBusy"
                    aria-label="关闭会话"
                  >
                    <span class="terminal-btn__icon">x</span>
                  </button>
                  <button
                    class="terminal-btn terminal-btn--icon terminal-btn--mac terminal-btn--min"
                    @click="handleMinimizePane(pane)"
                    :disabled="isBusy"
                    aria-label="最小化到列表"
                  >
                    <span class="terminal-btn__icon">-</span>
                  </button>
                  <button
                    class="terminal-btn terminal-btn--icon terminal-btn--mac terminal-btn--refresh"
                    @click="handleReconnectPane(pane)"
                    :disabled="isBusy || !pane.sessionId"
                    aria-label="重新连接终端"
                  >
                    <span class="terminal-btn__icon">r</span>
                  </button>
                </div>
                <div class="terminal-title" @dblclick="renamePaneSession(pane)">
                  {{ pane.sessionId ? getPaneTitle(pane) : "" }}
                </div>
                <div class="terminal-metrics">
                  <span>cols {{ pane.size.cols }}</span>
                  <span>rows {{ pane.size.rows }}</span>
                  <span>状态 {{ pane.status }}</span>
                </div>
              </div>
              <div class="terminal-body">
                <div
                  class="terminal"
                  :ref="(el) => setPaneContainer(pane.slot, el as HTMLDivElement | null)"
                ></div>
              </div>
              <div class="terminal-overlay" v-if="pane.busy">
                <div class="spinner"></div>
                <p>正在连接终端...</p>
              </div>
              <div class="terminal-empty" v-if="!pane.sessionId && sessionList.length > 0">
                拖动会话到此处
              </div>
            </div>
          </section>
        </section>
      </div>

      <n-drawer v-model:show="showRightDrawer" placement="right" :width="420">
        <n-drawer-content title="工作区面板">
          <n-collapse v-model:expanded-names="rightPanels">
            <n-collapse-item title="文件" name="files">
              <div class="drawer-section">
                <div class="drawer-actions">
                  <n-button size="small" @click="refreshFileRoot">刷新</n-button>
                  <n-button size="small" @click="triggerUpload">上传</n-button>
                  <n-button size="small" @click="downloadSelectedFile">下载</n-button>
                </div>
                <div class="path-breadcrumb">
                  <span class="path-breadcrumb__label">路径</span>
                  <div class="path-breadcrumb__trail">
                    <template v-for="(crumb, index) in fileBreadcrumbs" :key="crumb.path">
                      <button
                        class="path-breadcrumb__crumb"
                        :class="{ 'path-breadcrumb__crumb--active': index === fileBreadcrumbs.length - 1 }"
                        type="button"
                        @click="navigateToDir(crumb.path)"
                      >
                        {{ crumb.label }}
                      </button>
                      <span class="path-breadcrumb__sep" v-if="index < fileBreadcrumbs.length - 1">/</span>
                    </template>
                  </div>
                </div>
            <n-scrollbar style="max-height: 360px">
              <n-tree
                :data="fileTree"
                :on-load="handleLoadTree"
                :node-props="getFileNodeProps"
                :selected-keys="fileSelectedKey"
                @update:selected-keys="handleSelectTree"
              />
            </n-scrollbar>
                <input ref="fileInput" class="drawer-file-input" type="file" @change="handleUploadChange" />
              </div>
            </n-collapse-item>
            <n-collapse-item title="Git" name="git">
              <div class="drawer-section">
                <div class="drawer-actions">
                  <n-button size="small" @click="refreshGitStatus">刷新</n-button>
                </div>
                <n-scrollbar style="max-height: 220px">
                  <n-list>
                    <n-list-item
                      v-for="file in gitFiles"
                      :key="file.path"
                      class="git-item-card"
                      :class="{ 'git-item--active': gitSelected?.path === file.path }"
                      @click="selectGitFile(file)"
                      @dblclick="openGitDiffForFile(file)"
                    >
                      <div class="git-item">
                        <span
                          class="git-item__status"
                          :class="{
                            'git-item__status--added': file.status.includes('A') || file.status.includes('??'),
                            'git-item__status--deleted': file.status.includes('D'),
                            'git-item__status--modified': file.status.includes('M')
                          }"
                        >
                          {{ file.displayStatus }}
                        </span>
                        <span
                          class="git-item__path"
                          :class="{
                            'git-item__path--added': file.status.includes('A') || file.status.includes('??'),
                            'git-item__path--deleted': file.status.includes('D'),
                            'git-item__path--modified': file.status.includes('M')
                          }"
                        >
                          {{ file.display }}
                        </span>
                        <span class="git-item__stats" v-if="file.additions > 0 || file.deletions > 0">
                          <span class="git-item__stat git-item__stat--add" v-if="file.additions > 0">
                            +{{ file.additions }}
                          </span>
                          <span class="git-item__stat git-item__stat--del" v-if="file.deletions > 0">
                            -{{ file.deletions }}
                          </span>
                        </span>
                      </div>
                    </n-list-item>
                  </n-list>
                </n-scrollbar>
                <div class="drawer-hint" v-if="gitEmptyHint">{{ gitEmptyHint }}</div>
                <div class="git-diff-hint" v-if="gitDiff">已在弹窗中展示 diff</div>
              </div>
            </n-collapse-item>
          </n-collapse>
        </n-drawer-content>
      </n-drawer>

      <n-drawer v-model:show="showLogDrawer" placement="right" :width="520">
        <n-drawer-content title="后端日志">
          <div class="drawer-section log-panel">
            <div class="drawer-actions log-actions">
              <n-button size="small" @click="fetchBackendLogs" :loading="logLoading">刷新</n-button>
              <div class="log-switch">
                <span>自动滚动</span>
                <n-switch v-model:value="logAutoScroll" size="small" />
              </div>
            </div>
            <div class="log-hint">默认显示最后 100 行</div>
            <n-scrollbar ref="logScrollbar" style="max-height: 520px">
              <pre class="log-output" v-if="logText">{{ logText }}</pre>
              <div class="log-empty" v-else>暂无日志</div>
            </n-scrollbar>
          </div>
        </n-drawer-content>
      </n-drawer>

      <footer class="footer"></footer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed, nextTick, watch, h } from "vue";
import type { TreeOption } from "naive-ui";
import { useMessage, useDialog, NCode, NSwitch } from "naive-ui";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";

type FileTreeNode = TreeOption & {
  path: string;
  isDir: boolean;
  size?: number;
};

type GitFile = {
  path: string;
  status: string;
  display: string;
  additions: number;
  deletions: number;
  displayStatus: string;
};

const terminalContainers = ref<Record<number, HTMLDivElement | null>>({});
const socketRefs = ref<Record<number, WebSocket | null>>({});
const isBusy = ref(false);
const isConnecting = ref(false);
const appOnline = ref(typeof navigator !== "undefined" ? navigator.onLine : true);
const sessionList = ref<Array<{ id: string; name: string; last_active: string; display_index: number }>>([]);
const isSidebarCollapsed = ref(false);
const showRightDrawer = ref(false);
const showLogDrawer = ref(false);
const rightPanels = ref<string[]>(["files", "git"]);
const menu = ref<{ visible: boolean; x: number; y: number; session: { id: string; name: string } | null }>(
  { visible: false, x: 0, y: 0, session: null }
);
const layoutMode = ref<"single" | "vertical">("single");
const activePane = ref(0);
const hoverPane = ref<number | null>(null);
const panes = ref(
  [0, 1].map((slot) => ({
    slot,
    sessionId: "",
    title: "",
    term: null as Terminal | null,
    fitAddon: null as FitAddon | null,
    resizeObserver: null as ResizeObserver | null,
    wheelHandler: null as ((event: WheelEvent) => void) | null,
    status: "待命",
    size: { cols: 0, rows: 0 },
    busy: false
  }))
);
const fileRoot = ref("");
const fileTree = ref<FileTreeNode[]>([]);
const fileSelected = ref<FileTreeNode | null>(null);
const fileCurrentDir = ref(".");
const filePathInput = ref(".");
const fileInput = ref<HTMLInputElement | null>(null);
const gitFiles = ref<GitFile[]>([]);
const gitSelected = ref<GitFile | null>(null);
const gitDiff = ref("");
const gitEmptyHint = ref("");
const logScrollbar = ref<any>(null);
const logText = ref("");
const logLoading = ref(false);
const logAutoScroll = ref(false);
const message = useMessage();
const dialog = useDialog();

// getApiBaseURL 统一生成后端 API 基础地址（固定 8080）。
function getApiBaseURL() {
  const protocol = window.location.protocol;
  const host = window.location.hostname;
  return `${protocol}//${host}:8080`;
}

// apiURL 统一拼接后端 API 请求地址。
function apiURL(path: string, params?: URLSearchParams) {
  const base = getApiBaseURL();
  if (!params) {
    return `${base}${path}`;
  }
  const query = params.toString();
  return query ? `${base}${path}?${query}` : `${base}${path}`;
}

const headerStatus = computed(() => {
  if (panes.value.some((pane) => pane.busy)) {
    return "连接中";
  }
  if (appOnline.value) {
    return "在线";
  }
  return "离线";
});

const fileBreadcrumbs = computed(() => {
  const current = toRelativePath(fileCurrentDir.value);
  const parts = current.split("/").filter((part) => part.length > 0 && part !== ".");
  const rootLabel = fileRoot.value || ".";
  const crumbs = [{ label: rootLabel, path: "." }];
  let acc = "";
  parts.forEach((part) => {
    acc = acc ? `${acc}/${part}` : part;
    crumbs.push({ label: part, path: acc });
  });
  return crumbs;
});

const headerStatusClass = computed(() => {
  if (panes.value.some((pane) => pane.busy)) {
    return "status__dot--busy";
  }
  if (appOnline.value) {
    return "status__dot--online";
  }
  return "status__dot--offline";
});

const gridClass = computed(() => {
  if (layoutMode.value === "vertical") {
    return "terminal-grid--vertical";
  }
  return "terminal-grid--single";
});

const activeSessionId = computed(() => panes.value[activePane.value]?.sessionId || "");
const fileSelectedKey = computed(() => (fileSelected.value ? [fileSelected.value.key as string] : []));
const visiblePanes = computed(() => {
  if (layoutMode.value === "vertical") {
    return panes.value;
  }
  const active = panes.value[activePane.value];
  if (active && active.sessionId) {
    return [active];
  }
  const fallback = panes.value.find((pane) => pane.sessionId);
  return fallback ? [fallback] : [];
});

// fetchSessions 拉取会话列表。
async function fetchSessions() {
  try {
    const response = await fetch(apiURL("/api/sessions"));
    if (!response.ok) {
      return;
    }
    const payload = await response.json();
    sessionList.value = payload.sessions || [];
  } catch (error) {
    // ignore
  }
}

// ensureActiveSession 确保当前有激活会话。
function ensureActiveSession() {
  if (!activeSessionId.value) {
    message.warning("请先选择一个会话");
    return false;
  }
  return true;
}

// makeTreeNode 构造文件树节点。
function makeTreeNode(entry: { name: string; path: string; is_dir: boolean }) {
  const node: FileTreeNode = {
    label: entry.name,
    key: entry.path,
    path: entry.path,
    isDir: entry.is_dir,
    size: entry.size,
    isLeaf: !entry.is_dir
  };
  return node;
}

// fetchFileTree 拉取指定目录下的文件树节点。
async function fetchFileTree(path: string) {
  if (!ensureActiveSession()) {
    return { root: "", entries: [] as FileTreeNode[] };
  }
  const params = new URLSearchParams({
    session_id: activeSessionId.value,
    path
  });
  const response = await fetch(apiURL("/api/fs/tree", params));
  if (!response.ok) {
    throw new Error("无法读取目录");
  }
  const payload = await response.json();
  const entries = (payload.entries || []).map((entry: { name: string; path: string; is_dir: boolean }) =>
    makeTreeNode(entry)
  );
  return { root: payload.root || "", entries };
}

// normalizePath 规范化用户输入的路径。
function normalizePath(value: string) {
  const trimmed = value.trim();
  return trimmed === "" ? "." : trimmed;
}

// toRelativePath 将绝对路径转换为根目录下的相对路径。
function toRelativePath(value: string) {
  const trimmed = normalizePath(value);
  if (trimmed === "." || !trimmed.startsWith("/")) {
    return trimmed;
  }
  if (!fileRoot.value) {
    message.warning("暂不支持绝对路径");
    return ".";
  }
  const root = fileRoot.value.replace(/\/+$/, "");
  if (trimmed === root) {
    return ".";
  }
  if (trimmed.startsWith(`${root}/`)) {
    // 重要逻辑：将绝对路径转为相对路径，避免后端拒绝绝对路径。
    return normalizePath(trimmed.slice(root.length + 1));
  }
  message.warning("路径不在根目录下");
  return ".";
}

// navigateToDir 切换当前目录并刷新文件树。
async function navigateToDir(path: string) {
  const target = toRelativePath(path);
  try {
    const result = await fetchFileTree(target);
    fileRoot.value = result.root;
    fileTree.value = result.entries;
    fileSelected.value = null;
    fileCurrentDir.value = target;
    filePathInput.value = target;
  } catch (error) {
    message.error("目录加载失败");
  }
}

// refreshFileRoot 刷新文件树根目录。
async function refreshFileRoot() {
  await navigateToDir(".");
}

// handleLoadTree 懒加载目录节点。
async function handleLoadTree(node: TreeOption) {
  const target = node as FileTreeNode;
  if (!target.isDir) {
    return;
  }
  try {
    const result = await fetchFileTree(target.path);
    target.children = result.entries;
  } catch (error) {
    message.error("目录加载失败");
  }
}

// handleSelectTree 处理文件树节点选择。
function handleSelectTree(keys: Array<string | number>, options: Array<TreeOption | null>) {
  const selected = options[0] as FileTreeNode | null;
  fileSelected.value = selected || null;
}

// getFileNodeProps 绑定文件节点交互事件。
function getFileNodeProps({ option }: { option: TreeOption }) {
  const node = option as FileTreeNode;
  return {
    ondblclick: () => {
      if (node.isDir) {
        navigateToDir(node.path);
        return;
      }
      handlePreviewFile(node);
    }
  };
}

// applyPathInput 根据输入框内容切换目录。
function applyPathInput() {
  navigateToDir(filePathInput.value);
}

// goToParentDir 返回上一级目录。
function goToParentDir() {
  const current = normalizePath(fileCurrentDir.value);
  if (current === "." || current === "/") {
    navigateToDir(".");
    return;
  }
  const parts = current.split("/").filter((part) => part.length > 0 && part !== ".");
  if (parts.length === 0) {
    navigateToDir(".");
    return;
  }
  parts.pop();
  const parent = parts.length > 0 ? parts.join("/") : ".";
  navigateToDir(parent);
}

// triggerUpload 触发文件上传选择。
function triggerUpload() {
  if (!ensureActiveSession()) {
    return;
  }
  fileInput.value?.click();
}

// handleUploadChange 上传文件到当前目录。
async function handleUploadChange(event: Event) {
  const input = event.target as HTMLInputElement | null;
  const file = input?.files?.[0];
  if (!file) {
    return;
  }
  if (!ensureActiveSession()) {
    input.value = "";
    return;
  }
  const targetPath = fileSelected.value && fileSelected.value.isDir ? fileSelected.value.path : ".";
  const form = new FormData();
  form.append("file", file);
  const params = new URLSearchParams({
    session_id: activeSessionId.value,
    path: targetPath
  });
  try {
    const response = await fetch(apiURL("/api/fs/upload", params), {
      method: "POST",
      body: form
    });
    if (!response.ok) {
      throw new Error("upload failed");
    }
    await refreshFileRoot();
    message.success("上传成功");
  } catch (error) {
    message.error("上传失败");
  } finally {
    input.value = "";
  }
}

// downloadSelectedFile 下载选中的文件。
async function downloadSelectedFile() {
  if (!fileSelected.value || fileSelected.value.isDir) {
    message.warning("请选择一个文件");
    return;
  }
  if (!ensureActiveSession()) {
    return;
  }
  const params = new URLSearchParams({
    session_id: activeSessionId.value,
    path: fileSelected.value.path
  });
  const response = await fetch(apiURL("/api/fs/download", params));
  if (!response.ok) {
    message.error("下载失败");
    return;
  }
  const blob = await response.blob();
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = fileSelected.value.label as string;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(link.href);
}

// handlePreviewFile 处理文件预览。
async function handlePreviewFile(node: FileTreeNode) {
  if (!node || node.isDir) {
    return;
  }
  if (!ensureActiveSession()) {
    return;
  }
  if (typeof node.size === "number" && node.size > 1024 * 1024) {
    message.warning("文件超过 1MB，请下载查看");
    return;
  }
  const params = new URLSearchParams({
    session_id: activeSessionId.value,
    path: node.path
  });
  try {
    const response = await fetch(apiURL("/api/fs/read", params));
    if (!response.ok) {
      throw new Error("read failed");
    }
    const payload = await response.json();
    const text = payload.text || "";
    if (!text) {
      message.warning("文件内容为空");
      return;
    }
    openFilePreviewDialog({
      text,
      path: payload.path || node.path,
      size: payload.size || node.size || 0
    });
  } catch (error) {
    message.error("文件预览失败");
  }
}

// refreshGitStatus 刷新 Git 状态列表。
async function refreshGitStatus() {
  if (!ensureActiveSession()) {
    return;
  }
  const params = new URLSearchParams({
    session_id: activeSessionId.value
  });
  const response = await fetch(apiURL("/api/git/status", params));
  if (!response.ok) {
    gitFiles.value = [];
    gitDiff.value = "";
    gitEmptyHint.value = "当前目录不是 Git 仓库";
    return;
  }
  const payload = await response.json();
  const items = payload.items || [];
  gitFiles.value = items.map(
    (item: { path: string; status: string; orig_path?: string; additions?: number; deletions?: number }) => {
      const display = item.orig_path ? `${item.orig_path} -> ${item.path}` : item.path;
      const normalizedStatus = item.status || "";
      const displayStatus = normalizedStatus === "??" ? "新增" : normalizedStatus;
      return {
        path: item.path,
        status: normalizedStatus,
        display,
        additions: typeof item.additions === "number" ? item.additions : 0,
        deletions: typeof item.deletions === "number" ? item.deletions : 0,
        displayStatus
      };
    }
  );
  gitDiff.value = "";
  gitSelected.value = null;
  gitEmptyHint.value = gitFiles.value.length === 0 ? "暂无改动" : "";
}

// selectGitFile 选择 Git 文件（不自动打开 diff）。
function selectGitFile(file: GitFile) {
  gitSelected.value = file;
}

// openGitDiffForFile 双击 Git 文件后加载 diff 并弹窗展示。
async function openGitDiffForFile(file: GitFile) {
  if (!ensureActiveSession()) {
    return;
  }
  gitSelected.value = file;
  const params = new URLSearchParams({
    session_id: activeSessionId.value,
    path: file.path
  });
  const response = await fetch(apiURL("/api/git/diff", params));
  if (!response.ok) {
    gitDiff.value = "无法读取 diff";
    openGitDiffDialog(file.display, gitDiff.value);
    return;
  }
  const payload = await response.json();
  gitDiff.value = payload.diff || "";
  openGitDiffDialog(file.display, gitDiff.value);
}

// toggleSidebar 控制会话列表的收起与展开。
function toggleSidebar() {
  // 收起时同步关闭右键菜单，避免菜单悬浮在空白区域。
  if (!isSidebarCollapsed.value && menu.value.visible) {
    menu.value.visible = false;
  }
  isSidebarCollapsed.value = !isSidebarCollapsed.value;
}

// formatTime 格式化时间显示。
function formatTime(value: string) {
  if (!value) {
    return "未知";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

// updateAppOnline 同步浏览器网络状态到应用状态。
function updateAppOnline() {
  // 重要逻辑：右上角状态仅跟随全局网络状态，不绑定到任何会话。
  appOnline.value = navigator.onLine;
}

// isSessionInPane 判断会话是否已经在分栏中。
function isSessionInPane(id: string) {
  return panes.value.some((pane) => pane.sessionId === id);
}

// setPaneContainer 保存 pane 对应的容器引用。
function setPaneContainer(slot: number, el: HTMLDivElement | null) {
  terminalContainers.value[slot] = el;
}

// initTerminalForPane 初始化某个分栏的终端。
function initTerminalForPane(pane: typeof panes.value[number]) {
  if (!terminalContainers.value[pane.slot]) {
    return;
  }
  if (pane.term) {
    pane.term.dispose();
  }

  const terminal = new Terminal({
    fontFamily: "'IBM Plex Mono', ui-monospace",
    fontSize: 14,
    lineHeight: 1.4,
    cursorBlink: true,
    cursorStyle: "block",
    scrollback: 5000,
    scrollSensitivity: 1.5,
    smoothScrollback: true,
    theme: {
      background: "#07090b",
      foreground: "#e6f2ff",
      cursor: "#c6ff6f",
      selectionBackground: "#1c2a33",
      black: "#0b0f12",
      red: "#ff5f5f",
      green: "#b9ff6f",
      yellow: "#ffe27b",
      blue: "#5fc7ff",
      magenta: "#d18bff",
      cyan: "#63ffd4",
      white: "#f6fbff",
      brightBlack: "#1b242c",
      brightRed: "#ff8b8b",
      brightGreen: "#d4ff9a",
      brightYellow: "#ffeaa7",
      brightBlue: "#86d6ff",
      brightMagenta: "#e3b3ff",
      brightCyan: "#97ffe2",
      brightWhite: "#ffffff"
    }
  });

  const fitter = new FitAddon();
  terminal.loadAddon(fitter);
  terminal.open(terminalContainers.value[pane.slot] as HTMLDivElement);
  fitter.fit();

  terminal.onData((data) => {
    const socket = socketRefs.value[pane.slot];
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: "input", data }));
    }
  });

  pane.term = terminal;
  pane.fitAddon = fitter;

  pane.resizeObserver?.disconnect();
  const parent = (terminalContainers.value[pane.slot] as HTMLDivElement).parentElement;
  if (parent) {
    pane.resizeObserver = new ResizeObserver(() => {
      resizeTerminal(pane);
    });
    pane.resizeObserver.observe(parent);
  }

  pane.wheelHandler = (event: WheelEvent) => {
    if (!pane.term) {
      return;
    }
    event.preventDefault();
    const delta = event.deltaY > 0 ? 3 : -3;
    pane.term.scrollLines(delta);
  };
  terminalContainers.value[pane.slot]?.addEventListener("wheel", pane.wheelHandler, { passive: false });
}

// ensurePaneTerminalReady 确保当前分栏有可用的终端实例与容器绑定。
function ensurePaneTerminalReady(pane: typeof panes.value[number]) {
  const container = terminalContainers.value[pane.slot];
  if (!container) {
    return;
  }
  if (!pane.term) {
    initTerminalForPane(pane);
    return;
  }
  // 重要逻辑：容器重新渲染后重新初始化，避免终端绑在旧 DOM。
  if (container.childElementCount === 0) {
    pane.term.dispose();
    pane.term = null;
    pane.fitAddon = null;
    initTerminalForPane(pane);
  }
}

// resizeTerminal 调整指定分栏终端大小并通知后端。
function resizeTerminal(pane: typeof panes.value[number]) {
  if (!pane.fitAddon || !pane.term) {
    return;
  }
  pane.fitAddon.fit();
  const cols = pane.term.cols;
  const rows = pane.term.rows;
  pane.size = { cols, rows };

  const socket = socketRefs.value[pane.slot];
  if (socket && socket.readyState === WebSocket.OPEN) {
    // 重要逻辑：同步终端大小到后端 PTY。
    socket.send(JSON.stringify({ type: "resize", cols, rows }));
  }
}

// createSession 调用后端创建 PTY 会话。
async function createSession() {
  const response = await fetch(apiURL("/api/session"), { method: "POST" });
  if (!response.ok) {
    throw new Error("无法创建会话");
  }
  const payload = await response.json();
  return payload as { session_id: string; ws_url: string; name: string };
}

// connectWebSocket 连接指定会话到某个分栏。
async function connectWebSocket(sessionId: string, pane: typeof panes.value[number]) {
  pane.busy = true;
  pane.status = "连接中";

  // 重要逻辑：等待 DOM 渲染出终端容器后再初始化，避免首次连接黑屏。
  await nextTick();
  ensurePaneTerminalReady(pane);
  if (!pane.term) {
    initTerminalForPane(pane);
  }
  if (pane.term) {
    // 重要逻辑：切换/重连时先重置终端，避免缓存回放叠加。
    pane.term.reset();
  }

  const wsURL = `${getWSBaseURL()}?session_id=${sessionId}`;
  const oldSocket = socketRefs.value[pane.slot];
  if (oldSocket) {
    oldSocket.close();
  }

  return new Promise<void>((resolve, reject) => {
    const ws = new WebSocket(wsURL);
    socketRefs.value[pane.slot] = ws;

    ws.onopen = () => {
      pane.status = "在线";
      pane.busy = false;
      resizeTerminal(pane);
      requestAnimationFrame(() => {
        resizeTerminal(pane);
      });
      resolve();
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === "output" && pane.term) {
        pane.term.write(msg.data);
      }
      if (msg.type === "exit") {
        pane.status = "离线";
      }
    };

    ws.onerror = () => {
      pane.status = "离线";
      pane.busy = false;
      reject(new Error("ws error"));
    };

    ws.onclose = () => {
      if (pane.status !== "离线") {
        pane.status = "离线";
      }
    };
  });
}

// handleNewSession 新建会话并放入当前分栏。
async function handleNewSession() {
  try {
    isBusy.value = true;
    const payload = await createSession();
    const pane = panes.value[activePane.value];
    pane.sessionId = payload.session_id;
    pane.title = payload.name || "";
    await connectWebSocket(payload.session_id, pane);
    activePane.value = pane.slot;
    pane.term?.focus();
    await fetchSessions();
  } catch (error) {
    // ignore
  } finally {
    isBusy.value = false;
  }
}

// handleReconnect 重新连接当前分栏。
async function handleReconnectPane(pane: typeof panes.value[number]) {
  if (!pane.sessionId) {
    return;
  }
  try {
    await connectWebSocket(pane.sessionId, pane);
  } catch (error) {
    // ignore
  }
}

// handleClosePane 关闭指定分栏会话（空分栏也可清理）。
async function handleClosePane(pane: typeof panes.value[number]) {
  isBusy.value = true;
  try {
    if (pane.sessionId) {
      await fetch(apiURL("/api/session/close"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: pane.sessionId })
      });
    }
    clearPane(pane);
    await fetchSessions();
  } finally {
    isBusy.value = false;
  }
}

// handleMinimizePane 将分栏会话最小化回列表。
async function handleMinimizePane(pane: typeof panes.value[number]) {
  if (!pane.sessionId) {
    return;
  }
  detachPane(pane);
  await fetchSessions();
}

// clearPane 清理分栏资源。
function clearPane(pane: typeof panes.value[number]) {
  const socket = socketRefs.value[pane.slot];
  if (socket) {
    socket.close();
    socketRefs.value[pane.slot] = null;
  }
  pane.term?.clear();
  pane.sessionId = "";
  pane.title = "";
  pane.status = "待命";
  updateLayoutMode();
}

// detachPane 从界面移除会话，但不关闭后端会话。
function detachPane(pane: typeof panes.value[number]) {
  const socket = socketRefs.value[pane.slot];
  if (socket) {
    socket.close();
    socketRefs.value[pane.slot] = null;
  }
  pane.term?.clear();
  pane.sessionId = "";
  pane.title = "";
  pane.status = "待命";
  updateLayoutMode();
}

// handleConnectToActivePane 从列表连接到当前分栏。
async function handleConnectToActivePane(sessionId: string) {
  const existingPane = panes.value.find((pane) => pane.sessionId === sessionId);
  if (existingPane) {
    // 重要逻辑：已在分栏中则仅切换激活分栏，避免重复连接。
    activePane.value = existingPane.slot;
    existingPane.title = resolveSessionName(sessionId);
    updateLayoutMode();
    try {
      isConnecting.value = true;
      await nextTick();
      ensurePaneTerminalReady(existingPane);
      resizeTerminal(existingPane);
    } finally {
      isConnecting.value = false;
    }
    return;
  }
  try {
    isConnecting.value = true;
    const pane = panes.value[activePane.value];
    pane.sessionId = sessionId;
    pane.title = resolveSessionName(sessionId);
    await connectWebSocket(sessionId, pane);
  } finally {
    isConnecting.value = false;
  }
}

// resolveSessionName 根据 ID 找到名称。
function resolveSessionName(sessionId: string) {
  const index = sessionList.value.findIndex((item) => item.id === sessionId);
  if (index >= 0) {
    const found = sessionList.value[index];
    if (found && found.name) {
      return found.name;
    }
    if (found && found.display_index > 0) {
      return `会话 ${found.display_index}`;
    }
    return `会话 ${index + 1}`;
  }
  return "";
}

// handleDragStart 设置拖动数据。
function handleDragStart(event: DragEvent, sessionId: string) {
  event.dataTransfer?.setData("text/plain", sessionId);
}

// resolvePaneSlotFromEvent 根据鼠标位置计算分栏槽位。
function resolvePaneSlotFromEvent(event: DragEvent) {
  const grid = event.currentTarget as HTMLElement | null;
  if (!grid) {
    return 0;
  }
  const rect = grid.getBoundingClientRect();
  const isRight = event.clientX - rect.left > rect.width / 2;
  return isRight ? 1 : 0;
}

// handleGridDragOver 根据拖动位置决定分栏。
function handleGridDragOver(event: DragEvent) {
  const grid = event.currentTarget as HTMLElement;
  if (!grid) {
    return;
  }
  hoverPane.value = resolvePaneSlotFromEvent(event);
  if (layoutMode.value !== "vertical") {
    layoutMode.value = "vertical";
  }
}

// handleGridDragLeave 清理拖动状态。
function handleGridDragLeave(event: DragEvent) {
  const grid = event.currentTarget as HTMLElement | null;
  const nextTarget = event.relatedTarget as Node | null;
  if (grid && nextTarget && grid.contains(nextTarget)) {
    return;
  }
  hoverPane.value = null;
  updateLayoutMode();
}

// handleDropToGrid 处理拖动到空白区域的分栏放置。
async function handleDropToGrid(event: DragEvent) {
  const sessionId = event.dataTransfer?.getData("text/plain") || "";
  if (!sessionId) {
    return;
  }
  const slot = resolvePaneSlotFromEvent(event);
  await handleDropToPane(event, slot);
}

// handleDropToPane 拖动放置到分栏。
async function handleDropToPane(event: DragEvent, slot: number) {
  event.stopPropagation();
  const sessionId = event.dataTransfer?.getData("text/plain") || "";
  if (!sessionId) {
    return;
  }
  layoutMode.value = "vertical";
  const pane = panes.value[slot];
  const sourcePane = panes.value.find((item) => item.sessionId === sessionId) || null;
  try {
    isConnecting.value = true;
    if (sourcePane && sourcePane.slot !== slot) {
      if (pane.sessionId) {
        // 重要逻辑：两个分栏都已有会话时，交换会话位置，避免重复会话。
        const targetSessionId = pane.sessionId;
        const targetTitle = pane.title;
        pane.sessionId = sessionId;
        pane.title = resolveSessionName(sessionId);
        sourcePane.sessionId = targetSessionId;
        sourcePane.title = targetTitle || resolveSessionName(targetSessionId);
        activePane.value = slot;
        hoverPane.value = null;
        await connectWebSocket(pane.sessionId, pane);
        await connectWebSocket(sourcePane.sessionId, sourcePane);
        updateLayoutMode();
        return;
      }
      // 重要逻辑：目标分栏为空时，移动会话到目标分栏。
      sourcePane.sessionId = "";
      sourcePane.title = "";
      sourcePane.status = "待命";
      pane.sessionId = sessionId;
      pane.title = resolveSessionName(sessionId);
      activePane.value = slot;
      hoverPane.value = null;
      await connectWebSocket(sessionId, pane);
      updateLayoutMode();
      return;
    }
    if (pane.sessionId && pane.sessionId !== sessionId) {
      // 重要逻辑：拖入新会话替换当前分栏。
      pane.term?.clear();
    }
    pane.sessionId = sessionId;
    pane.title = resolveSessionName(sessionId);
    activePane.value = slot;
    hoverPane.value = null;
    await connectWebSocket(sessionId, pane);
    updateLayoutMode();
  } finally {
    isConnecting.value = false;
  }
}

// handleDropToList 拖动回列表移除分栏。
function handleDropToList(event: DragEvent) {
  const sessionId = event.dataTransfer?.getData("text/plain") || "";
  if (!sessionId) {
    return;
  }
  const pane = panes.value.find((item) => item.sessionId === sessionId);
  if (pane) {
    clearPane(pane);
  }
  updateLayoutMode();
}

// updateLayoutMode 根据会话数量调整布局。
function updateLayoutMode() {
  const count = panes.value.filter((pane) => pane.sessionId).length;
  if (count >= 2) {
    layoutMode.value = "vertical";
    return;
  }
  layoutMode.value = "single";
  activePane.value = 0;
}

// getPaneTitle 返回分栏标题，默认使用序号。
function getPaneTitle(pane: typeof panes.value[number]) {
  if (pane.title) {
    return pane.title;
  }
  return `会话 ${pane.slot + 1}`;
}

// getListTitle 列表展示标题。
function getListTitle(session: { id: string; name: string; display_index: number }) {
  if (session.name) {
    return session.name;
  }
  if (session.display_index > 0) {
    return `会话 ${session.display_index}`;
  }
  return "会话";
}

// openSessionMenu 打开会话右键菜单。
function openSessionMenu(event: MouseEvent, session: { id: string; name: string }) {
  event.stopPropagation();
  menu.value = { visible: true, x: event.clientX, y: event.clientY, session };
}

// closeMenu 关闭右键菜单。
function closeMenu() {
  menu.value = { visible: false, x: 0, y: 0, session: null };
}

// renameSession 重命名会话。
async function renameSession(session: { id: string; name: string } | null) {
  if (!session) {
    return;
  }
  const nextName = window.prompt("请输入会话名称", session.name || session.id);
  if (!nextName) {
    closeMenu();
    return;
  }
  await fetch(apiURL("/api/session/rename"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ session_id: session.id, name: nextName })
  });
  await fetchSessions();
  panes.value.forEach((pane) => {
    if (pane.sessionId === session.id) {
      pane.title = nextName;
    }
  });
  closeMenu();
}

// closeSession 关闭会话。
async function closeSession(session: { id: string; name: string } | null) {
  if (!session) {
    return;
  }
  await fetch(apiURL("/api/session/close"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ session_id: session.id })
  });
  const pane = panes.value.find((item) => item.sessionId === session.id);
  if (pane) {
    clearPane(pane);
  }
  await fetchSessions();
  closeMenu();
}

// renamePaneSession 双击标题重命名当前分栏。
async function renamePaneSession(pane: typeof panes.value[number]) {
  if (!pane.sessionId) {
    return;
  }
  await renameSession({ id: pane.sessionId, name: pane.title || pane.sessionId });
}

// getWSBaseURL 生成 WebSocket 基础地址。
function getWSBaseURL() {
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const host = window.location.hostname;
  return `${protocol}://${host}:8080/api/ws`;
}

// handleKeydown 处理快捷键关闭菜单。
function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    closeMenu();
  }
}

// openFilePreviewDialog 使用 Naive UI 弹窗展示文件内容。
function openFilePreviewDialog(payload: { text: string; path: string; size: number }) {
  dialog.create({
    title: "文件预览",
    style: "width: min(820px, 90vw)",
    content: () =>
      h("div", { class: "modal-content" }, [
        h("div", { class: "modal-subtitle" }, `${payload.path} (${formatSize(payload.size)})`),
        h("pre", { class: "modal-pre" }, payload.text)
      ])
  });
}

// openGitDiffDialog 使用 Naive UI 弹窗展示 diff。
function openGitDiffDialog(path: string, diff: string) {
  dialog.create({
    title: "Git Diff",
    style: "width: min(860px, 92vw)",
    content: () =>
      h("div", { class: "modal-content" }, [
        h("div", { class: "modal-subtitle" }, path),
        h(NCode, { code: diff, language: "diff", class: "git-diff-code" })
      ])
  });
}

// formatSize 格式化文件大小。
function formatSize(bytes: number) {
  if (!bytes || bytes < 0) {
    return "0 B";
  }
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// fetchBackendLogs 拉取后端日志内容。
async function fetchBackendLogs() {
  try {
    logLoading.value = true;
    const params = new URLSearchParams({ lines: "100" });
    const response = await fetch(apiURL("/api/logs/backend", params));
    if (!response.ok) {
      throw new Error("load failed");
    }
    const payload = await response.json();
    logText.value = payload.text || "";
    scrollLogToBottom();
  } catch (error) {
    logText.value = "";
    message.error("日志获取失败");
  } finally {
    logLoading.value = false;
  }
}

// scrollLogToBottom 自动滚动日志到末尾。
function scrollLogToBottom() {
  if (!logAutoScroll.value) {
    return;
  }
  nextTick(() => {
    logScrollbar.value?.scrollTo({ top: Number.MAX_SAFE_INTEGER });
  });
}

onMounted(async () => {
  updateAppOnline();
  window.addEventListener("online", updateAppOnline);
  window.addEventListener("offline", updateAppOnline);
  await nextTick();
  await fetchSessions();
  window.addEventListener("click", closeMenu);
  window.addEventListener("scroll", closeMenu, true);
  window.addEventListener("resize", closeMenu);
  window.addEventListener("keydown", handleKeydown);
});

watch(
  () => activeSessionId.value,
  async () => {
    if (activeSessionId.value) {
      return;
    }
    fileRoot.value = "";
    fileTree.value = [];
    gitFiles.value = [];
    gitDiff.value = "";
    gitEmptyHint.value = "";
    fileCurrentDir.value = ".";
    filePathInput.value = ".";
  }
);

watch(
  () => showRightDrawer.value,
  async (visible) => {
    if (!visible) {
      return;
    }
    await refreshFileRoot();
    await refreshGitStatus();
  }
);

watch(
  () => showLogDrawer.value,
  async (visible) => {
    if (!visible) {
      return;
    }
    await fetchBackendLogs();
  }
);

watch(
  () => logText.value,
  () => {
    scrollLogToBottom();
  }
);

watch([activePane, layoutMode], async () => {
  await nextTick();
  const pane = panes.value[activePane.value];
  if (!pane) {
    return;
  }
  ensurePaneTerminalReady(pane);
  resizeTerminal(pane);
});

onBeforeUnmount(() => {
  window.removeEventListener("online", updateAppOnline);
  window.removeEventListener("offline", updateAppOnline);
  window.removeEventListener("click", closeMenu);
  window.removeEventListener("scroll", closeMenu, true);
  window.removeEventListener("resize", closeMenu);
  window.removeEventListener("keydown", handleKeydown);
  panes.value.forEach((pane) => {
    pane.resizeObserver?.disconnect();
    if (pane.wheelHandler && terminalContainers.value[pane.slot]) {
      terminalContainers.value[pane.slot]?.removeEventListener("wheel", pane.wheelHandler);
    }
    pane.term?.dispose();
  });
});
</script>
