# Anywhere Code

在 Windows 上提供网页终端，终端进程运行在 WSL + tmux 中。

## 依赖

- Windows 10
- 已安装 WSL（Ubuntu 等）
- WSL 内安装 tmux
- Python 3.12（或本机可用的 Python 3）
- Node.js 18+

## 后端（FastAPI）

1. 复制环境变量示例

```bash
copy backend\example.env backend\.env
```

2. 设置账号密码

编辑 `backend/.env`：

```
ANYWHERE_AUTH_USERNAME=admin
ANYWHERE_AUTH_PASSWORD=admin
ANYWHERE_DATA_DIR=./data
ANYWHERE_MAX_SESSIONS=10
```

3. 安装依赖并启动

```bash
cd backend
python -m pip install -r requirements.txt
python -m uvicorn app.main:app --reload
```

> 如果使用 Poetry，可执行：
> `poetry install` + `poetry run uvicorn app.main:app --reload`

## 前端（Vue3 + Vite）

```bash
cd frontend
npm install
npm run dev
```

## WSL/tmux 说明

后端通过 `wsl.exe tmux` 调用 WSL 内的 tmux：

- 创建会话：`tmux new-session -d -s <id>`
- 附着会话：`tmux attach -t <id>`
- 列出会话：`tmux list-sessions`

确保 WSL 内的 `tmux` 可直接运行。
