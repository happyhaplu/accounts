#!/bin/sh
set -e

# ── Postgres host ─────────────────────────────────────────────────────────────
# "db" is the postgres service name in docker-compose.  Can be overridden via
# the DB_HOST env var in Coolify / docker-compose, but defaults to "db".
export DB_HOST="${DB_HOST:-db}"

# ── Backend URL ───────────────────────────────────────────────────────────────
# nginx and the Go API ALWAYS run inside this same combined container.
# Force 127.0.0.1 — MUST be exported so envsubst substitutes it.
# This overrides any BACKEND_URL Coolify may have injected from stale UI config.
export BACKEND_URL="http://127.0.0.1:8080"

# Substitute only ${BACKEND_URL} — leave all nginx variables ($host, $remote_addr, etc.) untouched.
# Template lives in /tmp so nginx:alpine's own entrypoint never touches it.
envsubst '${BACKEND_URL}' \
  < /tmp/nginx.conf.template \
  > /etc/nginx/conf.d/default.conf

# Validate config before starting — surfaces any syntax error in the logs.
nginx -t

# Start the Go API in the background
/app/server &

# Start nginx in the foreground (PID 1)
exec nginx -g "daemon off;"
