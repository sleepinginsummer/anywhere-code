# ttyd + tmux 使用手册（网页控制）

## 1. 启动（WSL 里执行）
建议固定一个 tmux 会话名，方便断线重连。
`
C:\Users\syy/.local/bin/ttyd -W -p 7681 tmux new -A -s main
/home/syy/.local/bin/ttyd -W -p 7681 tmux new -A -s main
`
说明：
- -W 开启可写模式（否则只读无法输入）
- -p 7681 指定网页端口
- 	mux new -A -s main：如果 main 会话存在则附着，不存在则新建

## 2. 浏览器访问
在局域网任意设备浏览器打开：
`
http://Windows主机IP:7681
`

## 3. 基本 tmux 操作（常用）
前缀键默认是 Ctrl+b。
- 新建窗口：Ctrl+b 然后 c
- 切换窗口：Ctrl+b 然后数字键
- 分屏（左右）：Ctrl+b 然后 %
- 分屏（上下）：Ctrl+b 然后 "
- 关闭当前窗格：Ctrl+b 然后 x
- 断开会话：Ctrl+b 然后 d

## 4. 停止服务
在 tmux 里退出后，ttyd 进程会结束。
若在后台运行过 ttyd，可用：
`
pkill ttyd
`

## 5. 常见问题
### 5.1 浏览器无法访问
- 确认 ttyd 正在运行
- 确认 Windows 防火墙已放行 7681 端口
- 确认访问的是 Windows 主机 IP（不是 WSL IP）

### 5.2 无法输入
- 确认启动命令包含 -W（可写模式）
- 点击页面后再输入
- 避免多个浏览器标签同时连接同一个 ttyd

### 5.3 中文输入问题
- 建议使用现代浏览器（Chrome/Edge）
- 若有输入法问题，先点一次网页终端后再切换输入法

## 6. 安全提醒
当前为内网无认证模式，不建议暴露到公网。
如需认证，可改用 ttyd 的 -c user:pass 参数或前置反向代理。


export http_proxy="http://127.0.0.1:7890"
export https_proxy="http://127.0.0.1:7890"
sudo -E apt update


npm i -g @openai/codex@latest