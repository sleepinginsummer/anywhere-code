#!/usr/bin/env bash
set -euo pipefail

# 重要逻辑：移除 NO_COLOR 并强制启用颜色相关变量后启动 codex。
env -u NO_COLOR \
  TERM=xterm-256color \
  COLORTERM=truecolor \
  CLICOLOR=1 \
  CLICOLOR_FORCE=1 \
  FORCE_COLOR=1 \
  codex
