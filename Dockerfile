# ─────────────────────────────────────────────────────────────────────────────
# Gour Accounts — Production Dockerfile
#
# Single Go binary serves API + Vue SPA.  Coolify's Traefik handles TLS.
# ─────────────────────────────────────────────────────────────────────────────

# ── Stage 1: Build Vue / Vite frontend ──────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /build-gour-accounts-frontend

COPY frontend/package*.json ./
RUN npm ci --prefer-offline

COPY frontend/ .

# ── CACHE BUSTER ────────────────────────────────────────────────────────────
# Docker BuildKit caches layers by content-hash.  Coolify shares the Docker
# daemon across ALL projects on the server, so a cached "npm run build" from
# another project with a similar Dockerfile structure can be served here.
# SOURCE_COMMIT changes on every git push → forces npm run build to always
# re-run.
ARG SOURCE_COMMIT=unknown
RUN echo "Building commit: ${SOURCE_COMMIT}" && \
    npm run build && \
    echo "=== Build verification ===" && \
    cat dist/index.html && \
    echo "" && \
    grep -q "Gour Accounts" dist/index.html || \
      (echo "FATAL: dist/index.html does NOT contain 'Gour Accounts' — stale Docker cache!" && exit 1)

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

COPY --from=backend-builder /build-gour-accounts-backend/server .
COPY --from=frontend-builder /build-gour-accounts-frontend/dist ./dist

# Final verification visible in Coolify deployment log
RUN echo "=== FINAL IMAGE ===" && cat dist/index.html && echo ""

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -sf http://localhost:3000/api/v1/health || exit 1

CMD ["./server"]
