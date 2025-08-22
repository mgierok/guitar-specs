# Guitar Specs

A Go web application for managing guitar specifications with a modern frontend stack.

## Prerequisites
- Go 1.25+
- PostgreSQL
- SSL certificate and private key files (HTTPS-only)
- Node.js 18+ and npm

## Setup

### 1. Environment Configuration
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

# Logging (optional)
# Available levels: debug, info, warn, error
# During startup/shutdown, full logging is always enabled
# This setting only affects runtime logging
LOG_LEVEL=info
```

### 2. SSL Certificates
Generate self-signed certificates for local development:
```bash
make ssl-gen
```

### 3. Frontend Dependencies
Install frontend dependencies:
```bash
make frontend-install
```

## Build and Run

### Build
```bash
# Build everything (frontend + Go binary)
make build

# Build frontend only
make frontend-build
```

### Run
```bash
make run
```

The application serves HTTPS on `HOST:PORT` (default `0.0.0.0:8443`).

## Frontend Tech Stack
- **HTMX** - Dynamic HTML updates
- **Alpine.js** - Lightweight JavaScript framework
- **Tailwind CSS** - Utility-first CSS framework
- **esbuild** - Fast JavaScript bundler

## Logging Configuration

The application uses a dual-logging approach:

- **Startup/Shutdown Logging**: Always enabled at INFO level for critical application lifecycle events
- **Runtime Logging**: Configurable via `LOG_LEVEL` environment variable

### Available Log Levels
- `debug` - Verbose logging for development and troubleshooting
- `info` - Standard logging level (default)
- `warn` - Only warnings and errors
- `error` - Only error messages

### Example Usage
```bash
# Set log level via environment variable
export LOG_LEVEL=debug
make run

# Or set in .env file
LOG_LEVEL=warn
```

## Database Management

### Schema Dump
```bash
pg_dump \
  -h localhost -p 5432 \
  -U guitar_specs_owner -d guitar_specs \
  --schema-only --no-owner \
  --file db/schema.sql
```

### Full Backup
```bash
pg_dump \
  -h localhost -p 5432 \
  -U guitar_specs_owner -d guitar_specs \
  --format=custom \
  --file backup_full.dump
```

### Restore
```bash
createdb -h localhost -p 5432 -U guitar_specs_owner guitar_specs_clean
psql -h localhost -p 5432 -U guitar_specs_owner -d guitar_specs_clean -f db/schema.sql
```

## Development
```bash
# Watch mode for frontend development
make frontend-dev

# Run tests
make test

# Check environment configuration
make env-check
```
