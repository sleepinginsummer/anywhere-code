# Repository Guidelines

## Project Structure & Module Organization
- `backend/`: Go HTTP/WebSocket 服务与会话管理逻辑（入口：`backend/main.go`）。
- `web/`: Vite + Vue 3 前端（入口：`web/src/main.ts`）。
- `ttyd/`: 终端后端组件源码与构建脚本（含 `ttyd/src/` 与 `ttyd/html/`）。
- `docs/`: 使用与安装文档。

## Build, Test, and Development Commands
- `cd backend && go run .`：启动后端服务（默认读取 `backend/config.go` 中的配置）。
- `backend/backend-run.sh`：后台启动后端（日志输出到 `backend/backend.out`）。
- `cd web && npm install`：安装前端依赖（首次必需）。
- `cd web && npm run dev`：启动前端开发服务器（Vite）。
- `cd web && npm run build`：构建前端静态产物。
- `cd web && npm run preview`：预览构建产物。

## Runbook (WSL)
- 所有启动操作在 WSL 内执行。
- 启动前先停止旧进程。
- 后端、前端均以后台方式启动。
- 后端启动：`cd backend && ./backend-run.sh`（日志：`backend/backend.out`）。
- 前端启动：`cd web && setsid -f bash -lc "npm run dev -- --host 0.0.0.0 --port 8001 --strictPort" > web-dev.out 2>&1`（日志：`web/web-dev.out`）。
- 停止后端：`pkill -f "backend/main"` 或 `pkill -f "backend-run.sh"`。
- 停止前端：`pkill -f "npm run dev"` 或 `pkill -f "vite"`。
- 后端清理旧进程：`pkill -f "anywhere-code-backend|go run \\.\" || true`（避免旧进程占用端口或路由异常）。

## Coding Style & Naming Conventions
- Go：遵循 `gofmt`，导出符号使用 `PascalCase`，未导出使用 `camelCase`。
- Vue/TS：组件文件使用 `PascalCase`（示例：`TerminalView.vue`），变量使用 `camelCase`。
- 统一使用 UTF-8 无 BOM 编码。
- 新增函数需添加函数级注释，关键逻辑加行内注释。

## Testing Guidelines
- 目前未发现独立测试目录或测试脚本。
- 如补充测试，建议 Go 测试文件命名为 `*_test.go`，前端测试放在 `web/` 下并与源码同层级或 `__tests__/` 中。

## Commit & Pull Request Guidelines
- 提交信息风格为 `type: summary`（示例：`feat: add session management APIs`）。
- 常见类型：`feat`、`docs`、`chore`。
- PR 应包含：问题描述、影响范围、验证方式（命令或截图）。

## Agent-Specific Instructions
- 不使用未授权的外部工具或脚本。
- 避免无关改动，保持改动最小化。
