# Hyttekos Development Makefile

.PHONY: help build run test clean docker-build docker-up docker-down migrate deps

# Default target
help:
	@echo "🏔️ Hyttekos Development Commands"
	@echo ""
	@echo "Development:"
	@echo "  dev          - Run application in development mode"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start development environment"
	@echo "  docker-down  - Stop development environment"
	@echo "  docker-logs  - View container logs"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps         - Download Go dependencies"
	@echo "  deps-update  - Update Go dependencies"
	@echo ""
	@echo "Deployment:"
	@echo "  deploy       - Deploy to fly.io"

# Development
dev: deps
	@echo "🚀 Starting Hyttekos in development mode..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "📝 Created .env file from .env.example - please configure it"; fi
	go run cmd/server/main.go

build:
	@echo "🔨 Building Hyttekos..."
	go build -o bin/hyttekos cmd/server/main.go

test:
	@echo "🧪 Running tests..."
	go test ./...

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Database
migrate-up:
	@echo "⬆️ Running database migrations..."
	@if [ -z "$(DATABASE_URL)" ]; then echo "❌ DATABASE_URL not set"; exit 1; fi
	migrate -path migrations -database $(DATABASE_URL) up

migrate-down:
	@echo "⬇️ Rolling back database migrations..."
	@if [ -z "$(DATABASE_URL)" ]; then echo "❌ DATABASE_URL not set"; exit 1; fi
	migrate -path migrations -database $(DATABASE_URL) down

# Docker
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t hyttekos .

docker-up:
	@echo "🐳 Starting development environment..."
	docker-compose up -d
	@echo "✅ Development environment started!"
	@echo "🌐 Application: http://localhost:8080"
	@echo "🗃️ Database: postgres://hyttekos:hyttekos@localhost:5432/hyttekos"

docker-down:
	@echo "🐳 Stopping development environment..."
	docker-compose down

docker-logs:
	@echo "📋 Viewing container logs..."
	docker-compose logs -f

# Dependencies
deps:
	@echo "📦 Downloading Go dependencies..."
	go mod download
	go mod tidy

deps-update:
	@echo "📦 Updating Go dependencies..."
	go get -u ./...
	go mod tidy

# Deployment
deploy:
	@echo "🚀 Deploying to fly.io..."
	./deploy.sh

# Database setup for local development
setup-db:
	@echo "🗃️ Setting up local database..."
	@if command -v psql >/dev/null 2>&1; then \
		createdb hyttekos || true; \
		psql -d hyttekos -c "CREATE USER hyttekos WITH ENCRYPTED PASSWORD 'hyttekos';" || true; \
		psql -d hyttekos -c "GRANT ALL PRIVILEGES ON DATABASE hyttekos TO hyttekos;" || true; \
		echo "✅ Database setup complete"; \
	else \
		echo "❌ PostgreSQL client (psql) not found. Please install PostgreSQL or use Docker."; \
	fi

# Quick start for development
start: docker-up
	@echo "⏳ Waiting for services to start..."
	@sleep 5
	@echo "🌐 Opening browser..."
	@if command -v open >/dev/null 2>&1; then open http://localhost:8080; fi
	@if command -v xdg-open >/dev/null 2>&1; then xdg-open http://localhost:8080; fi

# Stop all
stop: docker-down