#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

exec setsid -f bash -lc "go run ." > backend.out 2>&1
