#!/usr/bin/env bash
set -euo pipefail

(
  exec redis-server
) >/dev/null 2>&1 &

exec reflex -s -v -g '*.go' go run main.go