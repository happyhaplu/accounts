#!/bin/sh
set -e

# Default BACKEND_URL to localhost since both services run in this container.
# Coolify can override this if you ever split into two containers later.
export BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"

# Substitute only ${BACKEND_URL} — leave all nginx variables ($host, $remote_addr, etc.) untouched
envsubst '${BACKEND_URL}' \
  < /etc/nginx/templates/default.conf.template \
  > /etc/nginx/conf.d/default.conf

# Start the Go API in the background
/app/server &

# Start nginx in the foreground (PID 1)
exec nginx -g "daemon off;"
