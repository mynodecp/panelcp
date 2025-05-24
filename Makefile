# MyNodeCP Makefile
# Production-ready build and deployment automation

.PHONY: help build test clean install dev frontend backend docker deploy

# Variables
BINARY_NAME=mynodecp
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Directories
BACKEND_DIR=backend
FRONTEND_DIR=frontend
BUILD_DIR=build
DIST_DIR=dist

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
dev: ## Start development servers
	@echo "Starting MyNodeCP development environment..."
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start backend development server
	@echo "Starting backend development server..."
	@cd $(BACKEND_DIR) && $(GOCMD) run cmd/server/main.go

dev-frontend: ## Start frontend development server
	@echo "Starting frontend development server..."
	@cd $(FRONTEND_DIR) && npm run dev

# Build targets
build: clean build-backend build-frontend ## Build both backend and frontend
	@echo "Build completed successfully!"

build-backend: ## Build backend binary
	@echo "Building backend..."
	@mkdir -p $(BUILD_DIR)
	@cd $(BACKEND_DIR) && $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Backend built successfully!"

build-frontend: ## Build frontend for production
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && npm ci && npm run build
	@mkdir -p $(BUILD_DIR)/public
	@cp -r $(FRONTEND_DIR)/dist/* $(BUILD_DIR)/public/
	@echo "Frontend built successfully!"

# Test targets
test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	@echo "Running backend tests..."
	@cd $(BACKEND_DIR) && $(GOTEST) -v ./...

test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	@cd $(FRONTEND_DIR) && npm test

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@cd $(BACKEND_DIR) && $(GOTEST) -coverprofile=coverage.out ./...
	@cd $(BACKEND_DIR) && $(GOCMD) tool cover -html=coverage.out -o coverage.html

# Lint targets
lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Lint backend code
	@echo "Linting backend..."
	@cd $(BACKEND_DIR) && golangci-lint run

lint-frontend: ## Lint frontend code
	@echo "Linting frontend..."
	@cd $(FRONTEND_DIR) && npm run lint

# Dependencies
deps: deps-backend deps-frontend ## Install all dependencies

deps-backend: ## Install backend dependencies
	@echo "Installing backend dependencies..."
	@cd $(BACKEND_DIR) && $(GOMOD) download
	@cd $(BACKEND_DIR) && $(GOMOD) tidy

deps-frontend: ## Install frontend dependencies
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && npm ci

# Database targets
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@cd $(BACKEND_DIR) && $(GOCMD) run cmd/migrate/main.go

db-seed: ## Seed database with sample data
	@echo "Seeding database..."
	@cd $(BACKEND_DIR) && $(GOCMD) run cmd/seed/main.go

db-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	@cd $(BACKEND_DIR) && $(GOCMD) run cmd/reset/main.go

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t mynodecp:$(VERSION) .
	@docker tag mynodecp:$(VERSION) mynodecp:latest

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -d --name mynodecp -p 8080:8080 mynodecp:latest

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	@docker stop mynodecp || true
	@docker rm mynodecp || true

# Production targets
dist: build ## Create distribution package
	@echo "Creating distribution package..."
	@mkdir -p $(DIST_DIR)
	@cp -r $(BUILD_DIR)/* $(DIST_DIR)/
	@cp scripts/install.sh $(DIST_DIR)/
	@cp backend/configs/config.yaml $(DIST_DIR)/
	@tar -czf mynodecp-$(VERSION).tar.gz -C $(DIST_DIR) .
	@echo "Distribution package created: mynodecp-$(VERSION).tar.gz"

install: ## Install MyNodeCP on the system
	@echo "Installing MyNodeCP..."
	@sudo ./scripts/install.sh

install-fix: ## Fix incomplete installation
	@echo "Fixing MyNodeCP installation..."
	@if [ ! -d "/opt/mynodecp" ]; then \
		echo "MyNodeCP not installed. Run 'make install' first."; \
		exit 1; \
	fi
	@sudo cp -r . /opt/mynodecp/
	@cd /opt/mynodecp/backend && sudo -u mynodecp /usr/local/go/bin/go mod download
	@cd /opt/mynodecp/backend && sudo -u mynodecp /usr/local/go/bin/go build -o mynodecp-server cmd/server/main.go
	@cd /opt/mynodecp/frontend && sudo -u mynodecp npm install && sudo -u mynodecp npm run build
	@sudo chown -R mynodecp:mynodecp /opt/mynodecp
	@sudo chmod +x /opt/mynodecp/backend/mynodecp-server
	@sudo systemctl daemon-reload
	@sudo systemctl restart mynodecp
	@echo "Installation fixed successfully!"

deploy: dist ## Deploy to production server
	@echo "Deploying MyNodeCP..."
	@./scripts/deploy.sh

# Security targets
security-scan: ## Run security scans
	@echo "Running security scans..."
	@cd $(BACKEND_DIR) && gosec ./...
	@cd $(FRONTEND_DIR) && npm audit

# Performance targets
benchmark: ## Run performance benchmarks
	@echo "Running benchmarks..."
	@cd $(BACKEND_DIR) && $(GOTEST) -bench=. -benchmem ./...

load-test: ## Run load tests
	@echo "Running load tests..."
	@./scripts/load-test.sh

# Monitoring targets
metrics: ## Generate metrics report
	@echo "Generating metrics..."
	@./scripts/metrics.sh

# Documentation targets
docs: ## Generate documentation
	@echo "Generating documentation..."
	@cd $(BACKEND_DIR) && godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Cleanup targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f mynodecp-*.tar.gz
	@cd $(BACKEND_DIR) && $(GOCLEAN)
	@cd $(FRONTEND_DIR) && rm -rf dist node_modules/.cache

clean-all: clean ## Clean everything including dependencies
	@echo "Cleaning all artifacts and dependencies..."
	@cd $(FRONTEND_DIR) && rm -rf node_modules

# Git targets
tag: ## Create a new git tag
	@echo "Current version: $(VERSION)"
	@read -p "Enter new version: " NEW_VERSION; \
	git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
	git push origin $$NEW_VERSION

release: test lint build dist ## Create a release
	@echo "Creating release $(VERSION)..."
	@gh release create $(VERSION) mynodecp-$(VERSION).tar.gz --title "MyNodeCP $(VERSION)" --notes "Release $(VERSION)"

# Health check targets
health: ## Check system health
	@echo "Checking system health..."
	@curl -f http://localhost:8080/health || echo "Service not running"

status: ## Show service status
	@echo "MyNodeCP Service Status:"
	@systemctl status mynodecp || echo "Service not installed"

logs: ## Show service logs
	@echo "MyNodeCP Service Logs:"
	@journalctl -u mynodecp -f

# Development utilities
format: ## Format code
	@echo "Formatting code..."
	@cd $(BACKEND_DIR) && $(GOCMD) fmt ./...
	@cd $(FRONTEND_DIR) && npm run lint:fix

update: ## Update dependencies
	@echo "Updating dependencies..."
	@cd $(BACKEND_DIR) && $(GOGET) -u ./...
	@cd $(BACKEND_DIR) && $(GOMOD) tidy
	@cd $(FRONTEND_DIR) && npm update

# CI/CD targets
ci: deps lint test build ## Run CI pipeline
	@echo "CI pipeline completed successfully!"

cd: ci dist deploy ## Run CD pipeline
	@echo "CD pipeline completed successfully!"

# Quick start
quick-start: deps build ## Quick start for development
	@echo "MyNodeCP is ready for development!"
	@echo "Run 'make dev' to start development servers"

# Production deployment
production: ## Deploy to production
	@echo "Deploying to production..."
	@make ci
	@make dist
	@./scripts/deploy-production.sh
