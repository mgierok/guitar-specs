# Guitar Specs

A secure and performant Go web application with built-in HTTPS support, comprehensive security features, and production-ready performance optimisations.

## Features

### Security Features
- **HTTPS Enforcement**: Full HTTPS support with automatic HTTP to HTTPS redirection
- **Security Headers**: Comprehensive security headers including CSP, XSS protection, and HSTS
- **Rate Limiting**: IP-based rate limiting to prevent abuse and DoS attacks
- **Input Sanitisation**: Log sanitisation to prevent injection attacks
- **SSL Certificate Validation**: Automatic validation of certificate format, expiry, and key compatibility

### Performance Features
- **Static Asset Compression**: Pre-compressed Brotli and Gzip files with intelligent fallback
- **Asset Versioning**: Cache-busting URLs for static assets with long-lived caching
- **ETag Support**: HTTP caching with content-based ETags for dynamic content
- **Buffer Pooling**: Efficient memory management for template rendering using sync.Pool
- **Precompressed Assets**: Brotli and Gzip compression for maximum bandwidth savings

## Quick Start

### Prerequisites
- Go 1.25 or later
- SSL certificate and private key files (for HTTPS)

### Development Setup
```bash
# Clone and setup
git clone <repository>
cd guitar-specs

# Generate self-signed SSL certificates for local development
make ssl-gen

# Create environment configuration
make env-create

# Run the application
make run
```

### Running with HTTP (Development)
```bash
# Default configuration runs on HTTP port 8080
make run
```

### Running with HTTPS (Production)
```bash
# Option 1: Using .env file (recommended)
# Create .env.production file with your settings
echo "ENABLE_HTTPS=true" > .env.production
echo "SSL_CERT_FILE=ssl/localhost.crt" >> .env.production
echo "SSL_KEY_FILE=ssl/localhost.key" >> .env.production
echo "PORT=8443" >> .env.production
echo "HTTP_REDIRECT_PORT=8080" >> .env.production

# Run the application
make run

# Option 2: Using environment variables directly
export ENABLE_HTTPS=true
export SSL_CERT_FILE=ssl/localhost.crt
export SSL_KEY_FILE=ssl/localhost.key
export PORT=8443
export HTTP_REDIRECT_PORT=8080

# Run the application
make run
```

## Configuration

### Environment Files Priority

The application loads configuration from `.env` files with the following priority order:
1. **`.env`** (base configuration, lowest priority)
2. **`.env.[ENVIRONMENT]`** (environment-specific, e.g., `.env.production`)
3. **`.env.local`** (local overrides, does NOT override existing variables)

**Important**: `.env.local` only sets variables that are not already defined, it does not override values from `.env` or `.env.[ENVIRONMENT]` files.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server host address (0.0.0.0 for all interfaces) |
| `PORT` | `8080` | Server port number (8443 for HTTPS, 8080 for HTTP) |
| `ENV` | `development` | Environment name (development, production, staging) |
| `ENABLE_HTTPS` | `false` | Enable HTTPS mode |
| `SSL_CERT_FILE` | `""` | Path to SSL certificate file |
| `SSL_KEY_FILE` | `""` | Path to SSL private key file |
| `REDIRECT_HTTP` | `true` | Redirect HTTP to HTTPS (only when HTTPS enabled) |
| `HTTP_REDIRECT_PORT` | `8080` | Port for HTTP redirect server (8080 for development, 80 for production) |

#### Advanced Server Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `READ_TIMEOUT` | `10s` | Request read timeout |
| `WRITE_TIMEOUT` | `30s` | Response write timeout |
| `IDLE_TIMEOUT` | `60s` | Connection idle timeout |
| `READ_HEADER_TIMEOUT` | `5s` | Header read timeout |
| `MAX_HEADER_BYTES` | `1048576` | Maximum header size in bytes (1MB) |

#### Security Options
| Variable | Default | Description |
|----------|---------|-------------|
| `TRUSTED_PROXIES` | `127.0.0.1,::1` | Comma-separated list of trusted proxy IPs |
| `RATE_LIMIT` | `100` | Requests per minute per IP address |
| `RATE_LIMIT_WINDOW` | `1m` | Rate limit window (e.g., 1m, 5m, 1h) |

### Creating Environment Files

#### Development Configuration
```bash
# .env (base configuration)
HOST=127.0.0.1
PORT=8080
ENV=development
ENABLE_HTTPS=false
RATE_LIMIT=100
RATE_LIMIT_WINDOW=1m
```

#### Production Configuration
```bash
# .env.production
HOST=0.0.0.0
PORT=443
ENV=production
ENABLE_HTTPS=true
SSL_CERT_FILE=/etc/ssl/certs/app.crt
SSL_KEY_FILE=/etc/ssl/private/app.key
REDIRECT_HTTP=true
HTTP_REDIRECT_PORT=80
RATE_LIMIT=1000
RATE_LIMIT_WINDOW=1m
```

#### Local Overrides
```bash
# .env.local (gitignored, only sets undefined variables)
HOST=127.0.0.1
PORT=9000
```

### HTTPS Configuration

When `ENABLE_HTTPS=true`:
- The application runs on HTTPS using the specified certificate and key
- HTTP requests are automatically redirected to HTTPS (if `REDIRECT_HTTP=true`)
- HSTS headers are automatically added to enforce HTTPS usage
- All security headers are optimised for HTTPS
- SSL certificate validation ensures:
  - Certificate format is valid (PEM/DER)
  - Certificate is not expired
  - Certificate is not yet to be valid
  - Certificate expires within 30 days (warning)
  - Private key is compatible with certificate
  - RSA key size is at least 2048 bits

## Makefile Commands

### Development Commands
```bash
# Build the application
make build

# Run tests
make test

# Run the application (uses .env files)
make run

# Clean development files
make clean
```

### SSL Certificate Management
```bash
# Generate self-signed SSL certificates for local development
make ssl-gen

# Clean SSL certificates
make ssl-clean
```

### Environment Management
```bash
# Create .env files from templates
make env-create

# Check environment configuration
make env-check

# Clean environment files
make env-clean
```

### Docker Commands
```bash
# Build Docker image
make docker

# Run with Docker
docker run -p 8443:8443 -e ENABLE_HTTPS=true \
  -e SSL_CERT_FILE=/certs/cert.crt \
  -e SSL_KEY_FILE=/certs/key.key \
  guitar-specs
```

## SSL Certificate Setup

### Self-Signed Certificates (Development)
```bash
# Generate self-signed certificates
make ssl-gen

# This creates:
# - ssl/localhost.crt (certificate)
# - ssl/localhost.key (private key)
```

### Production Certificates
For production, use certificates from a trusted Certificate Authority:
- **Let's Encrypt** (free, automated)
- **Commercial CA** (paid, manual)
- **Internal CA** (enterprise)

### Certificate Requirements
- **Format**: PEM or DER
- **Key Type**: RSA (minimum 2048 bits)
- **Validity**: Not expired, not yet to be valid
- **Compatibility**: Certificate and private key must match

## Security Considerations

- **HTTPS Only**: In production, always use HTTPS with valid SSL certificates
- **HSTS**: HTTP Strict Transport Security is automatically enabled for HTTPS connections
- **Security Headers**: Comprehensive security headers protect against common web vulnerabilities:
  - Content Security Policy (CSP)
  - X-Frame-Options
  - X-Content-Type-Options
  - X-XSS-Protection
  - Referrer-Policy
  - Permissions-Policy
- **Rate Limiting**: Built-in rate limiting prevents abuse and DoS attacks
- **Input Validation**: All user inputs are sanitised to prevent injection attacks

## Performance Features

### Static Asset Optimisation
- **Precompression**: Brotli and Gzip compression for maximum bandwidth savings
- **Cache Busting**: Asset versioning ensures clients always get the latest version
- **Immutable Caching**: Long-lived cache headers for static assets
- **Intelligent Fallback**: Automatic selection of best compression format

### Dynamic Content Optimisation
- **ETag Caching**: Content-based ETags for efficient caching
- **Compression**: Automatic compression for HTML, CSS, JavaScript, and JSON
- **Buffer Pooling**: Efficient memory management for template rendering

## Monitoring and Logging

The application includes structured logging with comprehensive request details:
- **Request Information**: Method, path, status code, response time
- **Client Details**: IP address, user agent, trusted proxy handling
- **Security Events**: Rate limiting violations, SSL validation errors
- **Performance Metrics**: Asset versioning status, compression ratios

### Log Format
```json
{
  "time": "2025-08-16T00:06:49.056+02:00",
  "level": "INFO",
  "msg": "asset versions built successfully",
  "count": 2
}
```

## Production Deployment

### Prerequisites
1. **Valid SSL certificates** from trusted CA
2. **Production environment** configuration
3. **Proper port configuration** (80/443)
4. **Security hardening** (firewall, rate limiting)

### Deployment Steps
```bash
# 1. Generate production configuration
cp .env.example .env.production
# Edit .env.production with production values

# 2. Build application
make build

# 3. Deploy with proper SSL certificates
# 4. Configure reverse proxy if needed
# 5. Set up monitoring and logging
```

### Production Checklist
- [ ] SSL certificates from trusted CA
- [ ] Production environment configuration
- [ ] Proper port configuration (80/443)
- [ ] Security headers enabled
- [ ] Rate limiting configured
- [ ] Monitoring and logging setup
- [ ] Backup and recovery procedures

## Troubleshooting

### Common Issues

#### SSL Certificate Errors
```bash
# Check certificate format
openssl x509 -in ssl/localhost.crt -text -noout

# Check private key
openssl rsa -in ssl/localhost.key -check -noout

# Regenerate certificates if needed
make ssl-clean
make ssl-gen
```

#### Port Binding Errors
```bash
# Check if ports are in use
lsof -i :8443
lsof -i :8080

# Kill processes if needed
kill <PID>
```

#### Environment Configuration
```bash
# Check environment configuration
make env-check

# Verify .env file loading
cat .env
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[Add your license information here]
