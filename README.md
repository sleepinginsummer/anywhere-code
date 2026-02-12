# Anywhere Code

## 功能
- 提供基于 Web 的终端访问与会话管理能力
- 后端采用 Go 提供 HTTP/WebSocket 服务
- 前端使用 Vite + Vue 3 提供交互界面

## 环境配置
- 操作系统：Windows 10（通过 WSL 运行服务）
- 后端：Go（建议 1.20+）
- 前端：Node.js（建议 18+）与 npm
- 终端组件：`ttyd/` 目录内包含源码与构建脚本

## 项目启动
- 启动后端：
  - `cd backend && ./backend-run.sh`
  - 日志：`backend/backend.out`
- 启动前端（首次需安装依赖）：
  - `cd web && npm install`
  - `cd web && setsid -f bash -lc "npm run dev -- --host 0.0.0.0 --port 8001 --strictPort" > web-dev.out 2>&1`
  - 日志：`web/web-dev.out`
- 停止服务：
  - 后端：`pkill -f "backend/main"` 或 `pkill -f "backend-run.sh"`
  - 前端：`pkill -f "npm run dev"` 或 `pkill -f "vite"`
