#!/bin/sh
set -e

# ── Postgres host ─────────────────────────────────────────────────────────────
# "db" is the postgres service name in docker-compose.  Can be overridden via
# the DB_HOST env var in Coolify / docker-compose, but defaults to "db".
export DB_HOST="${DB_HOST:-db}"

# ── Backend URL ───────────────────────────────────────────────────────────────
# nginx and the Go API ALWAYS run inside this same combined container.
# Force 127.0.0.1 — ignore any external BACKEND_URL value (e.g. Coolify may
# set it to "http://backend:8080" which cannot resolve inside this container).
BACKEND_URL="http://127.0.0.1:8080"

# Substitute only ${BACKEND_URL} — leave all nginx variables ($host, $remote_addr, etc.) untouched
envsubst '${BACKEND_URL}' \
  < /etc/nginx/templates/default.conf.template \
  > /etc/nginx/conf.d/default.conf

# Start the Go API in the background
/app/server &

# Start nginx in the foreground (PID 1)
exec nginx -g "daemon off;"
