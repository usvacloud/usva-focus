#!/usr/bin/env bash
set -euo pipefail

(
  exec redis-server
) >/dev/null 2>&1 &

exec reflex -s -v \
  -r "\.go$" go run cmd/usva/usva.go daemon