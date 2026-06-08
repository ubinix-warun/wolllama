# ═══ Wolllama API — multi-stage Docker build ═══
# Build context: repository root
#    docker build -f api/Dockerfile -t wolllama-api .
#
# The Go workspace ties api/ + pkg/ together, and the binary embeds
# the React SPA from site/dist. This Dockerfile builds everything
# from source in two build stages, producing a minimal runtime image.

# ── Stage 1: Build React SPA ──────────────────────────────────
FROM node:22-alpine AS spa-builder
WORKDIR /src
COPY site/package.json site/package-lock.json* ./
RUN npm ci
COPY site/ ./
RUN npm run build
# Output: /src/dist/

# ── Stage 2: Build Go binary (CGO for SQLite) ─────────────────
FROM golang:1.23-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev sqlite
WORKDIR /src

# Copy workspace + module files first for dependency caching.
# All three go.work modules must be present even if we only build api.
COPY go.work go.work.sum* ./
COPY api/go.mod api/go.sum* api/
COPY cli/go.mod cli/go.sum* cli/
COPY pkg/go.mod pkg/go.sum* pkg/

# Download deps for all workspace modules (layer cache: cheap when unchanged).
RUN go work sync

# Copy Go source
COPY api/ api/
COPY pkg/ pkg/

# Copy built SPA into embed location
COPY --from=spa-builder /src/dist/ api/site/dist/

# Static build — no runtime linker dependencies
RUN CGO_ENABLED=1 \
    go build \
    -C api \
    -ldflags '-extldflags "-static"' \
    -o wolllama-api \
    .

# Preload a seed database with a preview model so the deployed
# service has content immediately (no cold-start empty page).
RUN sqlite3 /seed.db < api/seed.sql

# ── Stage 3: Runtime ──────────────────────────────────────────
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=go-builder /src/api/wolllama-api .
COPY --from=go-builder /seed.db wolllama.db

# Cloud Run sets PORT to the desired listen port.
# Default matches the binary's own default (8080).
EXPOSE 8080
CMD ["./wolllama-api"]
