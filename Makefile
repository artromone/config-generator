# Makefile
.PHONY: generate validate config clean

generate:
	@echo "🔨 Generating configuration files..."
	go run tools/configgen/main.go generate

validate:
	@echo "🔍 Validating configuration..."
	go run tools/configgen/main.go validate

config: generate validate
	@echo "✅ Configuration setup complete!"

clean:
	@echo "🧹 Cleaning generated files..."
	rm -f config/config.go .env.example .env.local

install:
	@echo "📦 Installing dependencies..."
	go mod tidy

setup: install generate
	@echo "🚀 Project setup complete!"
	@echo "Don't forget to fill in your .env.local file!"
