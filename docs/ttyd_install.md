# ttyd + tmux 安装手册（WSL1 / Ubuntu）

适用环境：Win10 + WSL1（Ubuntu），局域网内通过浏览器控制 tmux。

## 1. 前置条件
- WSL 已安装并可进入 Ubuntu
- Windows 能联网（用于下载 ttyd 二进制）
- tmux 已安装（若未安装可用 sudo apt-get install -y tmux）

## 2. 下载并安装 ttyd（二进制方式）
说明：由于 WSL 内部访问 GitHub 可能超时，采用 Windows 侧下载后复制到 WSL。

### 2.1 在 Windows 侧下载
在 PowerShell 运行：
`
curl.exe -L --max-time 120 -o C:\Users\syy\AppData\Local\Temp\ttyd.x86_64 https://github.com/tsl0922/ttyd/releases/download/1.7.7/ttyd.x86_64
`

### 2.2 复制到 WSL 临时目录
在 PowerShell 运行：
`
Copy-Item -Force C:\Users\syy\AppData\Local\Temp\ttyd.x86_64 \\wsl$\Ubuntu\tmp\ttyd.x86_64
`

### 2.3 在 WSL 中安装到用户目录
进入 WSL（Ubuntu）执行：
`
mkdir -p ~/.local/bin
mv /tmp/ttyd.x86_64 ~/.local/bin/ttyd
chmod +x ~/.local/bin/ttyd
`

### 2.4 配置 PATH（仅需一次）
`
grep -q  export PATH=\C:\Users\syy/.local/bin:\" ~/.bashrc || echo export PATH=\C:\Users\syy/.local/bin:\" >> ~/.bashrc
source ~/.bashrc
`

### 2.5 验证安装
`
C:\Users\syy/.local/bin/ttyd -v
`
输出示例：	tyd version 1.7.7-40e79c7

## 3. Windows 防火墙放行端口（管理员权限）
在以管理员身份运行的 PowerShell 执行：
`
New-NetFirewallRule -DisplayName ttyd 7681 -Direction Inbound -Action Allow -Protocol TCP -LocalPort 7681
`

## 4. 启动示例（重要：可写模式）
必须使用 -W，否则只读无法输入：
`
C:\Users\syy/.local/bin/ttyd -W -p 7681 tmux new -A -s main
`

## 5. 可选：tmux 安装
若 tmux 未安装：
`
sudo apt-get update
sudo apt-get install -y tmux
`

## 6. 卸载
`
rm -f ~/.local/bin/ttyd
`
如需移除 PATH：编辑 ~/.bashrc 删除添加的那一行。