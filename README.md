# Guitar Specs

A secure and performant Go web application with **HTTPS-only** operation, comprehensive security features, and production-ready performance optimisations.

## Features

### Security Features
- **HTTPS Only**: Application runs exclusively in HTTPS mode with SSL certificate validation
- **Security Headers**: Comprehensive security headers including CSP, XSS protection, and other security measures
- **Input Sanitisation**: Log sanitisation to prevent injection attacks
- **SSL Certificate Validation**: Automatic validation of certificate format, expiry, and key compatibility

### Performance Features
- **Static Asset Compression**: Pre-compressed Brotli and Gzip files with intelligent fallback
- **Asset Versioning**: Cache-busting URLs for static assets with long-lived caching
- **Buffer Pooling**: Efficient memory management for template rendering using sync.Pool
- **Precompressed Assets**: Brotli and Gzip compression for maximum bandwidth savings

## Quick Start

### Prerequisites
- Go 1.25 or later
- **SSL certificate and private key files (required)**

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

### Running the Application
```bash
# The application runs exclusively in HTTPS mode
# Create .env file with your SSL certificate paths
echo "SSL_CERT_FILE=ssl/localhost.crt" > .env
echo "SSL_KEY_FILE=ssl/localhost.key" >> .env
echo "PORT=8443" >> .env

# Run the application
make run
```

## Configuration

### Environment Files Priority

The application loads configuration from `.env` files with the following priority order:
1. **`.env`** (base configuration, lowest priority)
2. **`.env.[ENVIRONMENT]** (environment-specific, e.g., `.env.production`)
3. **`.env.local`** (local overrides, does NOT override existing variables)

**Important**: `.env.local` only sets variables that are not already defined, it does not override values from `.env` or `.env.[ENVIRONMENT]` files.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server host address (0.0.0.0 for all interfaces) |
| `PORT` | `8443` | Server port number (HTTPS only) |
| `ENV` | `development` | Environment name (development, production, staging) |
| `SSL_CERT_FILE` | **required** | Path to SSL certificate file |
| `SSL_KEY_FILE` | **required** | Path to SSL private key file |

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

### Creating Environment Files

#### Development Configuration
```bash
# .env (base configuration)
HOST=127.0.0.1
PORT=8443
ENV=development
SSL_CERT_FILE=ssl/localhost.crt
SSL_KEY_FILE=ssl/localhost.key
```

#### Production Configuration
```bash
# .env.production
HOST=0.0.0.0
PORT=443
ENV=production
SSL_CERT_FILE=/etc/ssl/certs/app.crt
SSL_KEY_FILE=/etc/ssl/private/app.key
```

#### Local Overrides
```bash
# .env.local (gitignored, only sets undefined variables)
HOST=127.0.0.1
PORT=9000
```

### HTTPS Configuration

The application runs exclusively in HTTPS mode:
- **SSL certificates are required** - the application will not start without valid certificates
- **Security headers are optimised** for HTTPS
- **SSL certificate validation ensures**:
  - Certificate format is valid (PEM/DER)
  - Certificate is not expired
  - Certificate is not yet to be valid
  - Certificate expires within 30 days (warning)
  - Private key is compatible with certificate
  - RSA key size is at least 2048 bits

**Note**: HTTP to HTTPS redirection and HSTS are handled by Cloudflare or your reverse proxy, not by the application itself.

## Makefile Commands

### Development Commands
```bash
# Build the application
make build

# Run tests
make test

# Run the HTTPS application (requires SSL certificates)
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
docker run -p 8443:8443 \
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

- **HTTPS Only**: The application runs exclusively in HTTPS mode
- **Security Headers**: Comprehensive security headers protect against common web vulnerabilities:
  - Content Security Policy (CSP)
  - X-Frame-Options
  - X-Content-Type-Options
  - X-XSS-Protection
  - Referrer-Policy
  - Permissions-Policy
- **Input Validation**: All user inputs are sanitised to prevent injection attacks
- **SSL Validation**: Strict SSL certificate validation prevents security issues
- **CDN Security**: HSTS and HTTP→HTTPS redirection handled by Cloudflare

## Performance Features

### Static Asset Optimisation
- **Precompression**: Brotli and Gzip compression for maximum bandwidth savings
- **Cache Busting**: Asset versioning ensures clients always get the latest version
- **Immutable Caching**: Long-lived cache headers for static assets
- **Intelligent Fallback**: Automatic selection of best compression format

### Dynamic Content Optimisation
- **Buffer Pooling**: Efficient memory management for template rendering

## Monitoring and Logging

The application includes structured logging with comprehensive request details:
- **Request Information**: Method, path, status code, response time
- **Client Details**: IP address, user agent, trusted proxy handling
- **Security Events**: SSL validation errors, security violations
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
3. **Proper port configuration** (443 for production)
4. **Security hardening** (firewall, CDN protection)
5. **Reverse proxy** (Cloudflare, nginx, etc.) for HTTP→HTTPS redirection

### Deployment Steps
```bash
# 1. Generate production configuration
cp .env.example .env.production
# Edit .env.production with production values

# 2. Build application
make build

# 3. Deploy with proper SSL certificates
# 4. Configure reverse proxy for HTTP→HTTPS redirection
# 5. Set up monitoring and logging
```

### Production Checklist
- [ ] SSL certificates from trusted CA
- [ ] Production environment configuration
- [ ] Proper port configuration (443)
- [ ] Security headers enabled
- [ ] Monitoring and logging setup
- [ ] Backup and recovery procedures
- [ ] Reverse proxy configured for HTTP→HTTPS redirection
- [ ] CDN configured for rate limiting and DDoS protection

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

#### Missing SSL Certificates
```bash
# The application requires SSL certificates to start
# Generate them first:
make ssl-gen

# Or specify existing certificates in .env:
echo "SSL_CERT_FILE=/path/to/cert.crt" >> .env
echo "SSL_KEY_FILE=/path/to/key.key" >> .env
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
