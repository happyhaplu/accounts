# ─────────────────────────────────────────────────────────────────────────────
# Stage 1: Build Vue / Vite frontend
# ─────────────────────────────────────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci --prefer-offline
COPY frontend/ .
RUN npm run build

# ─────────────────────────────────────────────────────────────────────────────
# Stage 2: Build Go backend
# ─────────────────────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS backend-builder

RUN apk --no-cache add git ca-certificates

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
# Do NOT pin GOARCH — let Go target the native arch of the build host
# so the binary runs on both amd64 and arm64 Coolify servers.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server .

# ─────────────────────────────────────────────────────────────────────────────
# Stage 3: Runtime — nginx (serves SPA) + Go binary (API)
# ─────────────────────────────────────────────────────────────────────────────
FROM nginx:1.25-alpine

# ca-certificates → HTTPS outbound (SMTP, Stripe)
# tzdata          → correct log timestamps
# gettext         → envsubst for nginx config templating
# wget            → Docker / Coolify health checks
RUN apk --no-cache add ca-certificates tzdata gettext wget

# Remove the default nginx config so our startup script is the ONLY source
# of truth for /etc/nginx/conf.d/default.conf — prevents any race where
# the stock config is briefly active before docker-start.sh runs.
RUN rm -f /etc/nginx/conf.d/default.conf

# Go API binary
COPY --from=backend-builder /app/server /app/server

# Vue SPA static files
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# Nginx config template (BACKEND_URL substituted at startup by docker-start.sh)
# Stored in /tmp — NOT in /etc/nginx/templates/ — so nginx:alpine's own
# entrypoint never processes it (avoiding a double-substitution race where
# it would substitute ${BACKEND_URL} with whatever is in the environment
# at that moment, before docker-start.sh forces it to 127.0.0.1).
COPY frontend/nginx.conf.template /tmp/nginx.conf.template

# Startup script: fills nginx template, starts Go API, then runs nginx
COPY docker-start.sh /docker-start.sh
RUN chmod +x /docker-start.sh

# nginx: 80  |  Go API: 8080 (internal, not exposed outside container)
EXPOSE 80

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:80/ || exit 1

CMD ["/docker-start.sh"]
