# SSL Certificate Management
ssl-gen:
	@echo "→ Generating self-signed SSL certificates for local development..."
	@mkdir -p ssl
	@openssl req -x509 -newkey rsa:4096 -keyout ssl/localhost.key -out ssl/localhost.crt -days 365 -nodes -subj "/C=PL/ST=Warsaw/L=Warsaw/O=LocalDev/OU=IT/CN=localhost"
	@echo "✓ SSL certificates generated in ssl/ directory"
	@echo "→ You can now run 'make run-https' to start with HTTPS"

ssl-trust:
	@echo "→ Adding SSL certificate to system trust store..."
	@echo "→ This will require your password to install the certificate"
	@sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ssl/localhost.crt
	@echo "✓ Certificate added to system trust store"
	@echo "→ You may need to restart your browser"

ssl-clean:
	@echo "→ Removing SSL certificates..."
	@rm -rf ssl/
	@echo "✓ SSL certificates removed"

env-clean:
	@echo "→ Removing .env files..."
	@rm -f .env .env.local .env.backup
	@echo "✓ .env files removed"

clean: ssl-clean env-clean
	@echo "→ All development files cleaned"

# Precompress static assets (gzip + brotli)
assets-precompress:
	@echo "→ Precompressing /web/static assets..."
	@find web/static -type f \
	  \( -name '*.js' -o -name '*.css' -o -name '*.svg' -o -name '*.html' -o -name '*.json' -o -name '*.wasm' \) \
	  -print0 | xargs -0 -I{} sh -c 'gzip -k -f -9 "{}"; brotli -f -q 11 "{}"'
	@echo "✓ Done."

run:
	go generate ./...
	go run ./cmd/web

run-https:
	@echo "→ Starting application with HTTPS..."
	@echo "→ Make sure you have .env file with HTTPS configuration"
	@echo "→ Run 'make ssl-gen' to generate self-signed certificates"
	@if [ ! -f .env ]; then echo "❌ .env file not found. Run 'make dev-setup' first."; exit 1; fi
	go run ./cmd/web

run-https-bin:
	@echo "→ Starting compiled application with HTTPS..."
	@echo "→ Make sure you have .env file with HTTPS configuration"
	@echo "→ Run 'make ssl-gen' to generate self-signed certificates"
	@if [ ! -f .env ]; then echo "❌ .env file not found. Run 'make dev-setup' first."; exit 1; fi
	@make build
	./bin/web

build:
	go generate ./...
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/web ./cmd/web

test:
	go generate ./...
	go test ./...

test-https:
	@echo "→ Testing HTTPS endpoint..."
	@echo "→ Make sure the application is running with 'make run-https'"
	@if [ ! -f .env ]; then echo "❌ .env file not found. Run 'make env-create' first."; exit 1; fi
	@curl -k -s -o /dev/null -w "Status: %{http_code}, Time: %{time_total}s\n" https://localhost:8443/ || echo "❌ HTTPS test failed - make sure the app is running"

env-check:
	@echo "→ Checking .env configuration..."
	@if [ ! -f .env ]; then echo "❌ .env file not found"; exit 1; fi
	@echo "✓ .env file exists"
	@echo "→ Current configuration:"
	@grep -E "^(HOST|PORT|ENV|ENABLE_HTTPS|SSL_CERT_FILE|SSL_KEY_FILE|REDIRECT_HTTP|HTTP_REDIRECT_PORT)=" .env | sort

env-create:
	@echo "→ Creating .env file for development..."
	@if [ -f .env ]; then echo "⚠️  .env file already exists. Backing up to .env.backup"; cp .env .env.backup; fi
	@echo "HOST=127.0.0.1" > .env
	@echo "PORT=8443" >> .env
	@echo "ENV=development" >> .env
	@echo "ENABLE_HTTPS=true" >> .env
	@echo "SSL_CERT_FILE=ssl/localhost.crt" >> .env
	@echo "SSL_KEY_FILE=ssl/localhost.key" >> .env
	@echo "REDIRECT_HTTP=false" >> .env
	@echo "HTTP_REDIRECT_PORT=8080" >> .env
	@echo "✓ .env file created with HTTPS configuration"
	@echo "→ You can now run 'make run-https'"

dev-setup:
	@echo "→ Setting up development environment..."
	@make ssl-gen
	@make env-create
	@echo "✓ Development environment ready!"
	@echo "→ Run 'make run-https' to start with HTTPS"

dev-setup-trusted:
	@echo "→ Setting up development environment with trusted certificates..."
	@make ssl-gen
	@make ssl-trust
	@echo "✓ Development environment ready with trusted certificates!"
	@echo "→ Run 'make run-https' to start with HTTPS"
	@echo "→ Note: You may need to restart your browser"

# Help
help:
	@echo "Available commands:"
	@echo "  run          - Start application with HTTP (development)"
	@echo "  run-https    - Start application with HTTPS (development)"
	@echo "  run-https-bin- Start compiled application with HTTPS"
	@echo "  build        - Build application binary"
	@echo "  test         - Run tests"
	@echo "  test-https   - Test HTTPS endpoint"
	@echo "  env-check    - Check .env configuration"
	@echo "  ssl-gen      - Generate self-signed SSL certificates"
	@echo "  ssl-trust    - Add SSL certificate to system trust store"
	@echo "  ssl-clean    - Remove SSL certificates"
	@echo "  env-create   - Create .env file with HTTPS configuration"
	@echo "  env-clean    - Remove .env files"
	@echo "  dev-setup    - Setup development environment (SSL + .env)"
	@echo "  dev-setup-trusted - Setup with trusted certificates"
	@echo "  clean        - Clean all development files"
	@echo "  assets-precompress - Precompress static assets"
	@echo "  help         - Show this help message"
