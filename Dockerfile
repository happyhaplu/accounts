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
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server .

# ─────────────────────────────────────────────────────────────────────────────
# Stage 3: Runtime — single Go binary serves API + Vue SPA
#
# No nginx.  Fiber serves static files from ./dist and falls back to
# index.html for Vue Router client-side routes.  Coolify's Traefik
# handles TLS termination and reverse-proxying.
# ─────────────────────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# Go binary
COPY --from=backend-builder /app/server .

# Vue SPA — served by Fiber's app.Static()
COPY --from=frontend-builder /app/dist ./dist

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -sf http://localhost:3000/api/v1/health || exit 1

CMD ["./server"]
