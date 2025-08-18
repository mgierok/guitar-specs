# Guitar Specs

Minimal instructions to set up, build, and run the application.

## Prerequisites
- Go 1.25+
- PostgreSQL reachable from the app
- SSL certificate and private key files (HTTPS-only)
- **Optional**: Node.js 18+ and npm for advanced frontend development

## 1) Setup
Create a `.env` file with required settings:
```bash
# Server
HOST=0.0.0.0
PORT=8443

# HTTPS (required)
SSL_CERT_FILE=ssl/localhost.crt
SSL_KEY_FILE=ssl/localhost.key

# Database (required)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=guitar_specs
DB_SSLMODE=disable
```

Generate self-signed certs for local development (optional):
```bash
make ssl-gen
```

**Optional**: Install frontend dependencies for advanced development:
```bash
make frontend-install
```

## 2) Build
```bash
# Build frontend assets and Go binary
make build

# Or build frontend separately
make frontend-build
```

## 3) Run
```bash
make run
```

The app serves HTTPS on `HOST:PORT` (default `0.0.0.0:8443`).

## Frontend Tech Stack
- **HTMX** - Dynamic HTML updates without JavaScript
- **Alpine.js** - Lightweight JavaScript framework for interactivity
- **Tailwind CSS** - Utility-first CSS framework
- **Heroicons** - Beautiful SVG icons
- **esbuild** - Ultra-fast JavaScript bundler

**Note**: The project includes a `build-frontend.sh` script that works without npm, making it easy to get started without installing Node.js dependencies.

## Development
```bash
# Build frontend assets (uses shell script if npm not available)
make frontend-build

# Advanced: Install npm dependencies and use esbuild + Tailwind
make frontend-install
make frontend-dev  # Watch mode for development
```

## Database dump and restore

Schema-only dump:
```bash
pg_dump \
  -h localhost -p 5432 \
  -U guitar_specs_owner -d guitar_specs \
  --schema-only --no-owner \
  --file db/schema.sql
```

Full database backup (custom format):
```bash
pg_dump \
  -h localhost -p 5432 \
  -U guitar_specs_owner -d guitar_specs \
  --format=custom \
  --file backup_full.dump
```

Restore from schema dump:
```bash
createdb -h localhost -p 5432 -U guitar_specs_owner guitar_specs_clean
psql     -h localhost -p 5432 -U guitar_specs_owner -d guitar_specs -f db/schema.sql
```
