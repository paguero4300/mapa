# Makefile for Navegar Project with Qodo Integration

.PHONY: help setup clean test coverage lint security format check-format vet staticcheck qodo-setup qodo-analyze docker-build docker-run

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Setup and installation
setup: ## Install dependencies and development tools
	@echo "🔧 Setting up development environment..."
	cd backend && go mod download && go mod verify
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install mvdan.cc/gofumpt@latest
	mkdir -p reports
	@echo "✅ Setup completed!"

qodo-setup: ## Run the Qodo setup script
	@echo "🚀 Running Qodo setup..."
	@if [ -f "scripts/qodo-setup.sh" ]; then \
		chmod +x scripts/qodo-setup.sh && ./scripts/qodo-setup.sh; \
	else \
		echo "❌ Setup script not found"; \
	fi

# Code quality and testing
test: ## Run all tests
	@echo "🧪 Running tests..."
	cd backend && go test -v ./...

coverage: ## Generate test coverage report
	@echo "📊 Generating coverage report..."
	cd backend && go test -coverprofile=../reports/coverage.out ./...
	cd backend && go tool cover -html=../reports/coverage.out -o ../reports/coverage.html
	@echo "📈 Coverage report generated: reports/coverage.html"

test-race: ## Run tests with race detection
	@echo "🏃 Running tests with race detection..."
	cd backend && go test -race ./...

# Code formatting and linting
format: ## Format Go code
	@echo "🎨 Formatting Go code..."
	cd backend && gofmt -s -w .
	@echo "✅ Code formatted!"

check-format: ## Check if Go code is properly formatted
	@echo "🔍 Checking Go code formatting..."
	@cd backend && if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ Code is not formatted properly:"; \
		gofmt -s -l .; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted"; \
	fi

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	cd backend && go vet ./...

staticcheck: ## Run staticcheck
	@echo "🔍 Running staticcheck..."
	cd backend && staticcheck ./...

lint: check-format vet staticcheck ## Run all linting tools

# Security scanning
security: ## Run security scan with gosec
	@echo "🔒 Running security scan..."
	gosec -fmt json -out reports/gosec-report.json ./backend/...
	@echo "🛡️ Security report generated: reports/gosec-report.json"

security-console: ## Run security scan with console output
	@echo "🔒 Running security scan..."
	gosec ./backend/...

# Qodo integration (when CLI is available)
qodo-analyze: ## Run Qodo analysis (requires Qodo CLI)
	@echo "🔍 Running Qodo analysis..."
	@if command -v qodo >/dev/null 2>&1; then \
		qodo analyze --config .qodo.yml --output reports/; \
	else \
		echo "❌ Qodo CLI not found. Please install Qodo CLI first."; \
		echo "💡 For now, running individual tools..."; \
		$(MAKE) lint test coverage security; \
	fi

# Combined quality checks
check: lint test coverage security ## Run all quality checks
	@echo "✅ All quality checks completed!"

check-ci: ## Run checks suitable for CI environment
	@echo "🤖 Running CI checks..."
	$(MAKE) check-format
	$(MAKE) vet
	$(MAKE) staticcheck
	$(MAKE) test-race
	$(MAKE) coverage
	$(MAKE) security
	@echo "✅ CI checks completed!"

# Build targets
build: ## Build the application
	@echo "🔨 Building application..."
	cd backend && go build -o ../bin/navegar .
	@echo "✅ Build completed: bin/navegar"

build-linux: ## Build for Linux
	@echo "🔨 Building for Linux..."
	cd backend && GOOS=linux GOARCH=amd64 go build -o ../bin/navegar-linux .

build-windows: ## Build for Windows
	@echo "🔨 Building for Windows..."
	cd backend && GOOS=windows GOARCH=amd64 go build -o ../bin/navegar.exe .

build-all: build build-linux build-windows ## Build for all platforms

# Docker targets
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t navegar:latest .

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --env-file .env navegar:latest

# Development targets
dev: ## Run development server
	@echo "🚀 Starting development server..."
	cd backend && go run .

dev-watch: ## Run development server with file watching (requires entr)
	@echo "🚀 Starting development server with file watching..."
	@if command -v entr >/dev/null 2>&1; then \
		find backend -name "*.go" | entr -r sh -c 'cd backend && go run .'; \
	else \
		echo "❌ entr not found. Install with: brew install entr (macOS) or apt install entr (Ubuntu)"; \
		echo "💡 Falling back to regular dev server..."; \
		$(MAKE) dev; \
	fi

# Cleanup targets
clean: ## Clean build artifacts and reports
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -rf reports/
	cd backend && go clean
	@echo "✅ Cleanup completed!"

clean-deps: ## Clean dependency cache
	@echo "🧹 Cleaning dependency cache..."
	cd backend && go clean -modcache
	@echo "✅ Dependency cache cleaned!"

# Git hooks
install-hooks: ## Install git hooks for code quality
	@echo "🪝 Installing git hooks..."
	@mkdir -p .git/hooks
	@cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
echo "🔍 Running pre-commit checks..."
make check-format || exit 1
make vet || exit 1
make test || exit 1
echo "✅ Pre-commit checks passed!"
EOF
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Git hooks installed!"

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	cd backend && go doc -all > ../docs/api.md
	@echo "📖 Documentation generated: docs/api.md"

# Environment setup
env-example: ## Copy .env.example to .env
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "📝 Created .env from .env.example"; \
		echo "💡 Please edit .env with your configuration"; \
	else \
		echo "⚠️ .env already exists"; \
	fi

# Quick start for new developers
quickstart: env-example setup install-hooks ## Quick setup for new developers
	@echo "🎉 Quick start completed!"
	@echo "💡 Next steps:"
	@echo "   1. Edit .env with your configuration"
	@echo "   2. Run 'make dev' to start the development server"
	@echo "   3. Run 'make check' to verify everything works"

# Show project status
status: ## Show project status and quality metrics
	@echo "📊 Project Status:"
	@echo "=================="
	@echo "📁 Project: Navegar (Traccar Login)"
	@echo "🔧 Language: Go $(shell cd backend && go version | awk '{print $$3}')"
	@echo "📦 Dependencies: $(shell cd backend && go list -m all | wc -l) modules"
	@if [ -f "reports/coverage.out" ]; then \
		echo "📈 Test Coverage: $(shell cd backend && go tool cover -func=../reports/coverage.out | grep total | awk '{print $$3}')"; \
	else \
		echo "📈 Test Coverage: Not available (run 'make coverage')"; \
	fi
	@echo "🔍 Quality Tools: gofmt, go vet, staticcheck, gosec"
	@echo "🤖 CI/CD: GitHub Actions configured"
	@echo "📋 Qodo: Configured (.qodo.yml)"
	@echo ""
	@echo "💡 Run 'make help' to see all available commands"