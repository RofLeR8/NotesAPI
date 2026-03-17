.PHONY: build run test clean docker-build docker-up docker-down docker-restart

# Build the application
build:
	go build -o bin/api ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t notes-api:latest .

# Start all services with docker-compose
docker-up:
	docker-compose up -d

# Stop all services
docker-down:
	docker-compose down

# Restart all services
docker-restart:
	docker-compose down && docker-compose up -d

# View logs
docker-logs:
	docker-compose logs -f notes_api

# Build and run
docker-run: docker-build docker-up

# Stop and remove containers + volumes
docker-clean:
	docker-compose down -v
