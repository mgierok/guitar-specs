# SSL Certificate Management
ssl-gen:
	@echo "→ Generating self-signed SSL certificates for local development..."
	@mkdir -p ssl
	@openssl req -x509 -newkey rsa:4096 -keyout ssl/localhost.key -out ssl/localhost.crt -days 365 -nodes -subj "/C=PL/ST=Warsaw/L=Warsaw/O=LocalDev/OU=IT/CN=localhost"
	@echo "✓ SSL certificates generated in ssl/ directory"

ssl-clean:
	@echo "→ Removing SSL certificates..."
	@rm -rf ssl/
	@echo "✓ SSL certificates removed"

run:
	go generate ./...
	go run ./cmd/web

build:
	go generate ./...
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/web ./cmd/web

test:
	go generate ./...
	go test ./...

env-check:
	@echo "→ Checking .env configuration..."
	@if [ ! -f .env ]; then echo "❌ .env file not found"; exit 1; fi
	@echo "✓ .env file exists"
	@echo "→ Current configuration:"
	@grep -E "^(HOST|PORT|ENV|SSL_CERT_FILE|SSL_KEY_FILE)=" .env | sort

env-clean:
	@echo "→ Removing .env files..."
	@rm -f .env .env.local .env.backup
	@echo "✓ .env files removed"

clean: ssl-clean env-clean
	@echo "→ All development files cleaned"

# Help
help:
	@echo "Available commands:"
	@echo "  run          - Start HTTPS application (requires SSL certificates)"
	@echo "  build        - Build application binary"
	@echo "  test         - Run tests"
	@echo "  env-check    - Check .env configuration"
	@echo "  ssl-gen      - Generate self-signed SSL certificates"
	@echo "  ssl-clean    - Remove SSL certificates"
	@echo "  env-clean    - Remove .env files"
	@echo "  clean        - Clean all development files"
	@echo "  help         - Show this help message"
