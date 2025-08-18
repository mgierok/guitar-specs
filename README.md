# Guitar Specs

Minimal instructions to set up, build, and run the application.

## Prerequisites
- Go 1.25+
- PostgreSQL reachable from the app
- SSL certificate and private key files (HTTPS-only)

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

## 2) Build
```bash
make build
```

## 3) Run
```bash
make run
```

The app serves HTTPS on `HOST:PORT` (default `0.0.0.0:8443`).

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
