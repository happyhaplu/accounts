# ─────────────────────────────────────────────────────────────────────────────
# Gour Accounts — Production Dockerfile
#
# Single Go binary serves API + Vue SPA.  Coolify's Traefik handles TLS.
# ─────────────────────────────────────────────────────────────────────────────

# ── Stage 1: Build Vue / Vite frontend ──────────────────────────────────────
FROM node:20-alpine AS frontend-builder

# Use a UNIQUE workdir so Docker cannot match cached layers from other
# projects (Warmup, Reach, etc.) that also use node:20-alpine + /app.
WORKDIR /build-gour-accounts-frontend

COPY frontend/package*.json ./
RUN npm ci --prefer-offline

COPY frontend/ .

# Verify source is correct, then build.  Output shows in Coolify deploy log.
RUN echo "=== Source check ===" && \
    grep '<title>' index.html && \
    npm run build && \
    echo "=== Build output ===" && \
    cat dist/index.html

# ── Stage 2: Build Go backend ──────────────────────────────────────────────
FROM golang:1.23-alpine AS backend-builder

RUN apk --no-cache add git ca-certificates

WORKDIR /build-gour-accounts-backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server .

# ── Stage 3: Runtime — single Go binary + Vue dist ─────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# Go binary
COPY --from=backend-builder /build-gour-accounts-backend/server .

# Vue SPA — served by Fiber's app.Get("/*") handler
COPY --from=frontend-builder /build-gour-accounts-frontend/dist ./dist

# Verify dist at build time (visible in Coolify deployment log)
RUN echo "=== FINAL IMAGE: dist/index.html ===" && cat dist/index.html

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -sf http://localhost:3000/api/v1/health || exit 1

CMD ["./server"]
