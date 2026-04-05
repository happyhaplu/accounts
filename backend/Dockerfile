# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# git is needed by some go modules; ca-certificates for HTTPS during build
RUN apk --no-cache add git ca-certificates

WORKDIR /app

# Cache dependency downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
# Do NOT pin GOARCH — let Go target the native architecture of the build host
# so the binary runs correctly on both amd64 and arm64 Coolify servers.
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server .

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates  → HTTPS outbound (SMTP relay, Stripe API)
# tzdata           → correct timestamps in logs
# curl             → Docker health check
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/server .

# Coolify / Docker runtime env vars supply all config.
# The app reads them directly; no .env file is needed inside the image.
EXPOSE 8080

CMD ["./server"]
