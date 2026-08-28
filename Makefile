.PHONY: run build test fmt migrate clean docker-up docker-down docker-build docker-registry docker-logs

# Default binary name and output directory
BINARY_NAME=flow
BIN_DIR=bin

# Run the Go application natively (backend folder)
run:
	cd backend && go run cmd/server/main.go

# Build the production executable binary natively
build:
	mkdir -p $(BIN_DIR)
	cd backend && go build -o ../$(BIN_DIR)/$(BINARY_NAME) cmd/server/main.go

# Run all unit tests with race detection
test:
	cd backend && go test -v -race ./...

# Run database migrations manually
migrate:
	cd backend && go run cmd/server/main.go -migrate

# Format all Go source files
fmt:
	cd backend && go fmt ./...

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)

# --- DOCKER COMPOSE TARGETS (TP2) ---

# Start full containerized stack (database, backend, frontend)
docker-up:
	docker compose up -d

# Build images and start containerized stack
docker-build:
	docker compose up -d --build

# Stop containerized stack
docker-down:
	docker compose down

# Stop containerized stack and delete persistent volumes
docker-down-v:
	docker compose down -v

# Start containerized stack using published registry images
docker-registry:
	docker compose -f docker-compose.registry.yml up -d

# View live logs
docker-logs:
	docker compose logs -f
