# ═══════════════════════════════════════════════════════════════
# Альфа Юнит-1 — Multi-stage Docker build
# Stage 1: build Go binary
# Stage 2: minimal runtime (Alpine)
# ═══════════════════════════════════════════════════════════════

# ── Stage 1: Builder ─────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Install build tools (needed for modernc.org/sqlite pure Go build)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Cache module downloads separately from source code changes
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

# Build a static binary — CGO_ENABLED=0 is safe with modernc.org/sqlite
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.buildVersion=$(date +%Y%m%d)" \
    -trimpath \
    -o /app/alfa1 \
    .

# ── Stage 2: Runtime ─────────────────────────────────────────
FROM alpine:3.20

# Security: run as non-root user
RUN addgroup -S alfa && adduser -S -G alfa alfa

# CA certs for outgoing HTTPS calls (Yandex Maps iframe, CDN, etc.)
RUN apk add --no-cache ca-certificates tzdata

# Set timezone to Moscow
ENV TZ=Europe/Moscow

WORKDIR /app

# Copy binary and nothing else — templates and static files are embedded
COPY --from=builder /app/alfa1 .

# Ensure data directory exists with correct ownership
RUN mkdir -p /app/data && chown -R alfa:alfa /app

USER alfa

# Expose HTTP port (actual port configured via ENV)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:${PORT:-8080}/ > /dev/null || exit 1

# Default environment (can be overridden in docker-compose or .env)
ENV PORT=8080
ENV DB_PATH=/app/data/alfa1.db
ENV APP_ENV=production

ENTRYPOINT ["/app/alfa1"]
