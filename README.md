# Guitar Specs

A secure and performant Go web application with built-in HTTPS support.

## Features

### Security Features
- **HTTPS Enforcement**: Full HTTPS support with automatic HTTP to HTTPS redirection
- **Security Headers**: Comprehensive security headers including CSP, XSS protection, and HSTS
- **Rate Limiting**: IP-based rate limiting to prevent abuse
- **Input Sanitisation**: Log sanitisation to prevent injection attacks

### Performance Features
- **Static Asset Compression**: Pre-compressed Brotli and Gzip files with intelligent fallback
- **Asset Versioning**: Cache-busting URLs for static assets with long-lived caching
- **ETag Support**: HTTP caching with content-based ETags
- **Buffer Pooling**: Efficient memory management for template rendering

## Quick Start

### Prerequisites
- Go 1.25 or later
- SSL certificate and private key files (for HTTPS)

### Running with HTTP (Development)
```bash
make run
```

### Running with HTTPS (Production)
```bash
# Option 1: Using .env file (recommended)
# Create .env.production file with your settings
echo "ENABLE_HTTPS=true" > .env.production
echo "SSL_CERT_FILE=/path/to/certificate.crt" >> .env.production
echo "SSL_KEY_FILE=/path/to/private.key" >> .env.production
echo "PORT=8443" >> .env.production

# Run the application
make run

# Option 2: Using environment variables directly
export ENABLE_HTTPS=true
export SSL_CERT_FILE=/path/to/certificate.crt
export SSL_KEY_FILE=/path/to/private.key
export PORT=8443

# Run the application
make run
```

## Configuration

### Environment Variables

The application supports loading configuration from `.env` files with the following priority order:
1. **`.env.local`** (highest priority, should be gitignored)
2. **`.env.[ENVIRONMENT]`** (environment-specific, e.g., `.env.production`)
3. **`.env`** (default, lowest priority)

Environment variables can also be set directly in the shell.

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server host address |
| `PORT` | `8080` | Server port number |
| `ENV` | `development` | Environment name |
| `ENABLE_HTTPS` | `false` | Enable HTTPS mode |
| `SSL_CERT_FILE` | `""` | Path to SSL certificate file |
| `SSL_KEY_FILE` | `""` | Path to SSL private key file |
| `REDIRECT_HTTP` | `true` | Redirect HTTP to HTTPS (when HTTPS enabled) |

### Creating .env Files

Create a `.env` file in your project root:

```bash
# .env (for development)
HOST=0.0.0.0
PORT=8080
ENV=development
ENABLE_HTTPS=false

# .env.production (for production)
HOST=0.0.0.0
PORT=8443
ENV=production
ENABLE_HTTPS=true
SSL_CERT_FILE=/etc/ssl/certs/app.crt
SSL_KEY_FILE=/etc/ssl/private/app.key
REDIRECT_HTTP=true

# .env.local (for local overrides, gitignored)
HOST=127.0.0.1
PORT=9000
ENV=development
```

### HTTPS Configuration

When `ENABLE_HTTPS=true`:
- The application runs on HTTPS using the specified certificate and key
- HTTP requests are automatically redirected to HTTPS (if `REDIRECT_HTTP=true`)
- HSTS headers are automatically added to enforce HTTPS usage
- All security headers are optimised for HTTPS

## Building and Running

```bash
# Build the application
make build

# Run tests
make test

# Run the application
make run

# Build Docker image
make docker
```

## Docker

The application includes a multi-stage Dockerfile for production builds.

```bash
# Build Docker image
make docker

# Run with Docker
docker run -p 8443:8443 -e ENABLE_HTTPS=true \
  -e SSL_CERT_FILE=/certs/cert.crt \
  -e SSL_KEY_FILE=/certs/key.key \
  guitar-specs
```

## Security Considerations

- **HTTPS Only**: In production, always use HTTPS with valid SSL certificates
- **HSTS**: HTTP Strict Transport Security is automatically enabled for HTTPS connections
- **Security Headers**: Comprehensive security headers protect against common web vulnerabilities
- **Rate Limiting**: Built-in rate limiting prevents abuse and DoS attacks

## Monitoring

The application includes structured logging with request details:
- Request method, path, and status
- Response time and client IP
- User agent information
- Sanitised paths to prevent log injection
