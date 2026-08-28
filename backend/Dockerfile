# ---- Stage 1: Build (SDK) ----
FROM golang:1.24-alpine AS build

WORKDIR /app

# Cache dependency download layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and migrations
COPY . .

# Build static executable stripped of symbols
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/server \
    cmd/server/main.go

# ---- Stage 2: Runtime (Minimal) ----
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
