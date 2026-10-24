# ---- Stage 1: Build (SDK) ----
FROM golang:1.24-alpine AS build

WORKDIR /app

# Cache dependency download layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code, migrations and coverage gate script
COPY . .

# GIT_COMMIT is injected by CI (or compose) so /api/health can report which SHA is running.
ARG GIT_COMMIT=unknown

# Build static executable stripped of symbols
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X flow/internal/version.Commit=${GIT_COMMIT}" \
    -o /app/server \
    cmd/server/main.go

# ---- Stage 2: tests (reuses SDK + sources from build) ----
FROM build AS test
# Threshold is statement coverage on *service.go (business logic).
# Go has no native branch coverage — documented in decisiones.md.
ENV COVERAGE_THRESHOLD=70
ENV OUT_DIR=/out
ENTRYPOINT ["/bin/sh", "/app/scripts/check-coverage.sh"]

# ---- Stage 3: Runtime (Minimal) ----
FROM alpine:3.20 AS final

WORKDIR /app

# Install root CA certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Create unprivileged application user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy compiled binary from build stage
COPY --from=build /app/server /app/server

# Change ownership
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/server"]
