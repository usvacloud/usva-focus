#!/usr/bin/env bash
set -euo pipefail

(
  exec redis-server
) >/dev/null 2>&1 &

while true; do
  nc -z 127.0.0.1 6379 && break
  echo "waiting for redis"
  sleep 0.1
done
echo "redis ok, starting usva"

exec /app/usvad