.PHONY: help generate validate config clean install setup run test

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

generate: ## Generate configuration files from templates
	@echo "🔨 Generating configuration files..."
	@go run tools/configgen/main.go generate

validate: ## Validate configuration against templates
	@echo "🔍 Validating configuration..."
	@go run tools/configgen/main.go validate

clean: ## Clean generated files
	@echo "🧹 Cleaning generated files..."
	@rm -f config/config.go .env.example
	@echo "Generated files cleaned (keeping .env.local)"

install: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go mod download

setup: install generate ## Complete project setup
	@echo "🚀 Project setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Copy .env.example to .env.local: cp .env.example .env.local"
	@echo "2. Fill in your environment variables in .env.local"
	@echo "3. Run the example: make run"

run: ## Run the example application
	@echo "🏃 Running example application..."
	@go run main.go

.DEFAULT_GOAL := help
