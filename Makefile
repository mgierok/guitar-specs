# Frontend Build
frontend-install:
	@echo "→ Installing frontend dependencies..."
	@npm install
	@echo "✓ Frontend dependencies installed"
	@echo "→ Updating Browserslist database..."
	@if command -v npm >/dev/null 2>&1; then \
		npx update-browserslist-db@latest; \
		echo "✓ Browserslist database updated"; \
	else \
		echo "⚠️  npm not found, cannot update Browserslist"; \
	fi

frontend-update-browserslist:
	@echo "→ Updating Browserslist database..."
	@if command -v npm >/dev/null 2>&1; then \
		npx update-browserslist-db@latest; \
		echo "✓ Browserslist database updated"; \
	else \
		echo "⚠️  npm not found, cannot update Browserslist"; \
	fi

frontend-dev:
	@echo "→ Starting frontend development (esbuild + Tailwind watch)..."
	@npm run dev & npm run watch:css

frontend-build:
	@echo "→ Building frontend assets..."
	@if command -v npm >/dev/null 2>&1 && [ -f package.json ]; then \
		npm run build; \
	else \
		echo "❌ npm not found, cannot build frontend assets"; \
		exit 1; \
	fi
	@echo "✓ Frontend assets built"

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
	go run ./cmd/web

build: frontend-build
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/web ./cmd/web

test:
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

# Maintenance
maintenance: frontend-update-browserslist
	@echo "→ Frontend maintenance completed"

# Help
help:
	@echo "Available commands:"
	@echo "  frontend-install - Install frontend dependencies (npm install)"
	@echo "  frontend-update-browserslist - Update Browserslist database"
	@echo "  frontend-dev     - Start frontend development server"
	@echo "  frontend-build   - Build frontend assets for production"
	@echo "  run              - Start HTTPS application (requires SSL certificates)"
	@echo "  build            - Build application binary and frontend assets"
	@echo "  test             - Run tests"
	@echo "  env-check        - Check .env configuration"
	@echo "  ssl-gen          - Generate self-signed SSL certificates"
	@echo "  ssl-clean        - Remove SSL certificates"
	@echo "  env-clean        - Remove .env files"
	@echo "  clean            - Clean all development files"
	@echo "  maintenance      - Update frontend dependencies (Browserslist)"
	@echo "  help             - Show this help message"
