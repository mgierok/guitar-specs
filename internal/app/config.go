package app

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration settings.
// It provides a centralised way to manage environment-specific configuration
// with sensible defaults for development and production environments.
type Config struct {
	Host         string // Server host address (default: 0.0.0.0 for all interfaces)
	Port         string // Server port number (default: 8080)
	Env          string // Environment name (default: development)
	EnableHTTPS  bool   // Whether to enable HTTPS (default: false)
	CertFile     string // Path to SSL certificate file (default: "")
	KeyFile      string // Path to SSL private key file (default: "")
	RedirectHTTP bool   // Whether to redirect HTTP to HTTPS (default: true when HTTPS enabled)

	// Advanced configuration options
	ReadTimeout       time.Duration // Request read timeout (default: 10s)
	WriteTimeout      time.Duration // Response write timeout (default: 30s)
	IdleTimeout       time.Duration // Connection idle timeout (default: 60s)
	ReadHeaderTimeout time.Duration // Header read timeout (default: 5s)
	MaxHeaderBytes    int           // Maximum header size in bytes (default: 1MB)

	// Security options
	TrustedProxies  []string      // List of trusted proxy IPs for RealIP middleware
	RateLimit       int           // Requests per minute per IP (default: 100)
	RateLimitWindow time.Duration // Rate limit window (default: 1 minute)

	// HTTP redirect server port (for redirecting HTTP to HTTPS)
	HTTPRedirectPort int // Port for HTTP redirect server (default: 8080)
}

// LoadConfig loads configuration from environment variables with sensible defaults.
// This function ensures the application has valid configuration even when
// environment variables are not explicitly set.
// It automatically loads .env files before reading environment variables.
func LoadConfig() Config {
	// Load .env files first to populate environment variables
	LoadEnvFiles()

	enableHTTPS := getenv("ENABLE_HTTPS", "false") == "true"
	redirectHTTP := getenv("REDIRECT_HTTP", "true") == "true"

	// Validate logical configuration consistency
	if redirectHTTP && !enableHTTPS {
		// HTTP redirect only makes sense when HTTPS is enabled
		redirectHTTP = false
	}

	// Set default port based on HTTPS configuration
	defaultPort := "8080"
	if enableHTTPS {
		defaultPort = "8443"
	}

	cfg := Config{
		Host:         getenv("HOST", "0.0.0.0"),    // Bind to all network interfaces
		Port:         getenv("PORT", defaultPort),  // Port based on HTTPS configuration
		Env:          getenv("ENV", "development"), // Default to development mode
		EnableHTTPS:  enableHTTPS,                  // Enable HTTPS if ENABLE_HTTPS=true
		CertFile:     getenv("SSL_CERT_FILE", ""),  // SSL certificate file path
		KeyFile:      getenv("SSL_KEY_FILE", ""),   // SSL private key file path
		RedirectHTTP: redirectHTTP,                 // Redirect HTTP to HTTPS (only when HTTPS enabled)

		// Advanced configuration options
		ReadTimeout:       getDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:      getDuration("WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:       getDuration("IDLE_TIMEOUT", 60*time.Second),
		ReadHeaderTimeout: getDuration("READ_HEADER_TIMEOUT", 5*time.Second),
		MaxHeaderBytes:    getInt("MAX_HEADER_BYTES", 1<<20), // 1MB

		// Security options
		TrustedProxies:  getStringSlice("TRUSTED_PROXIES", []string{"127.0.0.1", "::1"}),
		RateLimit:       getInt("RATE_LIMIT", 100),
		RateLimitWindow: getDuration("RATE_LIMIT_WINDOW", time.Minute),

		// HTTP redirect server configuration
		HTTPRedirectPort: getInt("HTTP_REDIRECT_PORT", 8080), // Use port 8080 for development
	}

	return cfg
}

// Addr returns the formatted address string for the HTTP server.
// This combines the host and port into a format suitable for net.Listen.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// AddrHTTPS returns the formatted address string for the HTTPS server.
// This is typically the same as Addr() but can be different if needed.
func (c Config) AddrHTTPS() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// AddrHTTP returns the formatted address string for the HTTP redirect server.
// This uses the configurable HTTPRedirectPort instead of hardcoded port 80.
func (c Config) AddrHTTP() string {
	// Fallback to port 80 if HTTPRedirectPort is zero to preserve legacy behaviour
	port := c.HTTPRedirectPort
	if port == 0 {
		port = 80
	}
	return fmt.Sprintf("%s:%d", c.Host, port)
}

// getenv retrieves an environment variable with a fallback default value.
// This helper function ensures configuration is always valid and prevents
// the application from failing due to missing environment variables.
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// ValidateHTTPS ensures HTTPS configuration is valid when enabled.
// This function checks that certificate and key files exist and are readable.
func (c Config) ValidateHTTPS() error {
	if !c.EnableHTTPS {
		return nil // HTTPS not enabled, no validation needed
	}

	if c.CertFile == "" {
		return fmt.Errorf("HTTPS enabled but SSL_CERT_FILE not specified")
	}

	if c.KeyFile == "" {
		return fmt.Errorf("HTTPS enabled but SSL_KEY_FILE not specified")
	}

	// Check if certificate file exists and is readable
	if _, err := os.Stat(c.CertFile); os.IsNotExist(err) {
		return fmt.Errorf("SSL certificate file not found: %s", c.CertFile)
	}

	// Check if private key file exists and is readable
	if _, err := os.Stat(c.KeyFile); os.IsNotExist(err) {
		return fmt.Errorf("SSL private key file not found: %s", c.KeyFile)
	}

	// Validate certificate format and expiry
	if err := c.validateCertificate(); err != nil {
		return fmt.Errorf("SSL certificate validation failed: %w", err)
	}

	// Validate certificate-key compatibility
	if err := c.validateCertificateKeyPair(); err != nil {
		return fmt.Errorf("SSL certificate-key pair validation failed: %w", err)
	}

	// Validate logical consistency
	if c.RedirectHTTP && !c.EnableHTTPS {
		return fmt.Errorf("HTTP redirect enabled but HTTPS is disabled - this configuration is invalid")
	}

	return nil
}

// validateCertificate checks certificate format and expiry
// This prevents runtime errors from malformed or expired certificates
func (c Config) validateCertificate() error {
	// Read and parse the certificate file
	certData, err := os.ReadFile(c.CertFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse the certificate - handle both PEM and DER formats
	var cert *x509.Certificate

	// Try to parse as PEM first (most common for self-signed certificates)
	if block, _ := pem.Decode(certData); block != nil {
		// PEM format detected
		if block.Type != "CERTIFICATE" {
			return fmt.Errorf("PEM block is not a certificate (type: %s)", block.Type)
		}
		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse PEM certificate: %w", err)
		}
	} else {
		// Try to parse as raw DER (binary format)
		cert, err = x509.ParseCertificate(certData)
		if err != nil {
			return fmt.Errorf("failed to parse certificate (tried PEM and DER): %w", err)
		}
	}

	// Check if certificate is expired
	now := time.Now()
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate expired on %s", cert.NotAfter.Format("2006-01-02 15:04:05"))
	}

	// Check if certificate is not yet valid
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate not valid until %s", cert.NotBefore.Format("2006-01-02 15:04:05"))
	}

	// Check if certificate is close to expiry (within 30 days)
	expiryThreshold := 30 * 24 * time.Hour
	if time.Until(cert.NotAfter) < expiryThreshold {
		return fmt.Errorf("certificate expires soon on %s (within 30 days)", cert.NotAfter.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// validateCertificateKeyPair checks if certificate and private key are compatible
// This ensures the application can actually use the certificate-key pair
func (c Config) validateCertificateKeyPair() error {
	// Read certificate and private key
	certData, err := os.ReadFile(c.CertFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	keyData, err := os.ReadFile(c.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	// Parse the certificate - handle both PEM and DER formats
	var cert *x509.Certificate
	if block, _ := pem.Decode(certData); block != nil {
		// PEM format detected
		if block.Type != "CERTIFICATE" {
			return fmt.Errorf("PEM block is not a certificate (type: %s)", block.Type)
		}
		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse PEM certificate: %w", err)
		}
	} else {
		// Try to parse as raw DER (binary format)
		cert, err = x509.ParseCertificate(certData)
		if err != nil {
			return fmt.Errorf("failed to parse certificate (tried PEM and DER): %w", err)
		}
	}

	// Parse the private key - handle both PEM and DER formats
	var privateKey interface{}

	// Try to parse as PEM first
	if block, _ := pem.Decode(keyData); block != nil {
		// PEM format detected
		switch block.Type {
		case "RSA PRIVATE KEY":
			// PKCS1 format
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse PEM RSA private key: %w", err)
			}
			privateKey = key
		case "PRIVATE KEY":
			// PKCS8 format
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse PEM PKCS8 private key: %w", err)
			}
			privateKey = key
		default:
			return fmt.Errorf("unsupported PEM block type for private key: %s", block.Type)
		}
	} else {
		// Try to parse as raw DER - try PKCS1 first, then PKCS8
		key, err := x509.ParsePKCS1PrivateKey(keyData)
		if err != nil {
			// Try PKCS8 format
			privateKey, err = x509.ParsePKCS8PrivateKey(keyData)
			if err != nil {
				return fmt.Errorf("failed to parse private key (tried PEM and DER): %w", err)
			}
		} else {
			privateKey = key
		}
	}

	// Check if it's an RSA key
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not RSA format")
	}

	// Verify RSA key size is reasonable
	if rsaKey.N.BitLen() < 2048 {
		return fmt.Errorf("RSA key size is too small: %d bits (minimum 2048 recommended)", rsaKey.N.BitLen())
	}

	// Verify certificate and key are compatible by checking public key
	if !reflect.DeepEqual(cert.PublicKey, rsaKey.Public()) {
		return fmt.Errorf("certificate and private key are not compatible")
	}

	return nil
}

// getDuration retrieves a duration from environment variable with fallback default.
func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// getInt retrieves an integer from environment variable with fallback default.
func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

// getStringSlice retrieves a string slice from environment variable with fallback default.
// Expected format: "ip1,ip2,ip3" (comma-separated)
func getStringSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return def
}
