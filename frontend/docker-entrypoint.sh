#!/bin/sh
set -e

# Substitute ONLY ${BACKEND_URL} — nginx's own variables ($host, $remote_addr,
# etc.) are passed as literal strings and must NOT be touched by envsubst.
envsubst '${BACKEND_URL}' \
  < /etc/nginx/templates/default.conf.template \
  > /etc/nginx/conf.d/default.conf

exec "$@"
