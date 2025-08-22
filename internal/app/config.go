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
	Host string // Server host address (default: 0.0.0.0 for all interfaces)
	Port string // Server port number (default: 8443 for HTTPS)
	Env  string // Environment name (default: development)

	// SSL Configuration (required for HTTPS)
	CertFile string // Path to SSL certificate file
	KeyFile  string // SSL private key file path

	// Database configuration (split parameters)
	DBHost     string // PostgreSQL host
	DBPort     string // PostgreSQL port (default: 5432)
	DBUser     string // PostgreSQL user
	DBPassword string // PostgreSQL password
	DBName     string // PostgreSQL database name
	DBSSLMode  string // sslmode (disable, require, verify-ca, verify-full)

	// Advanced configuration options
	ReadTimeout       time.Duration // Request read timeout (default: 10s)
	WriteTimeout      time.Duration // Response write timeout (default: 30s)
	IdleTimeout       time.Duration // Connection idle timeout (default: 60s)
	ReadHeaderTimeout time.Duration // Header read timeout (default: 5s)
	MaxHeaderBytes    int           // Maximum header size in bytes (1MB)

	// Security options
	TrustedProxies []string // List of trusted proxy IPs for RealIP middleware

	// Logging configuration
	LogLevel string // Log level for runtime (default: info)
}

// LoadConfig loads configuration from environment variables with sensible defaults.
// This function ensures the application has valid configuration even when
// environment variables are not explicitly set.
// It automatically loads .env file before reading environment variables.
func LoadConfig() Config {
	// Load .env file first to populate environment variables
	LoadEnvFile()

	cfg := Config{
		Host: getenv("HOST", "0.0.0.0"),    // Bind to all network interfaces
		Port: getenv("PORT", "8443"),       // Default to HTTPS port
		Env:  getenv("ENV", "development"), // Default to development mode

		// SSL Configuration
		CertFile: getenv("SSL_CERT_FILE", ""), // SSL certificate file path
		KeyFile:  getenv("SSL_KEY_FILE", ""),  // SSL private key file path

		// Database (split parameters)
		DBHost:     getenv("DB_HOST", ""),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", ""),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBName:     getenv("DB_NAME", ""),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),

		// Advanced configuration options
		ReadTimeout:       getDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:      getDuration("WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:       getDuration("IDLE_TIMEOUT", 60*time.Second),
		ReadHeaderTimeout: getDuration("READ_HEADER_TIMEOUT", 5*time.Second),
		MaxHeaderBytes:    getInt("MAX_HEADER_BYTES", 1<<20), // 1MB

		// Security options
		TrustedProxies: getStringSlice("TRUSTED_PROXIES", []string{"127.0.0.1", "::1"}),

		// Logging configuration
		LogLevel: getenv("LOG_LEVEL", "info"),
	}

	return cfg
}

// Addr returns the formatted address string for the HTTPS server.
// This combines the host and port into a format suitable for net.Listen.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
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

// ValidateHTTPS ensures HTTPS configuration is valid.
// This function checks that certificate and key files exist, are readable, and are in PEM format.
func (c Config) ValidateHTTPS() error {
	if c.CertFile == "" {
		return fmt.Errorf("SSL_CERT_FILE not specified")
	}

	if c.KeyFile == "" {
		return fmt.Errorf("SSL_KEY_FILE not specified")
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

	// Validate certificate-key compatibility (PEM format only)
	if err := c.validateCertificateKeyPair(); err != nil {
		return fmt.Errorf("SSL certificate-key pair validation failed: %w", err)
	}

	return nil
}

// validateCertificate checks certificate format and expiry
// This prevents runtime errors from malformed or expired certificates
// Only PEM format is supported
func (c Config) validateCertificate() error {
	// Read and parse the certificate file
	certData, err := os.ReadFile(c.CertFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse the certificate - PEM format only
	block, _ := pem.Decode(certData)
	if block == nil {
		return fmt.Errorf("certificate file is not in PEM format")
	}

	if block.Type != "CERTIFICATE" {
		return fmt.Errorf("PEM block is not a certificate (type: %s)", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse PEM certificate: %w", err)
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
// Only PEM format is supported
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

	// Parse the certificate - PEM format only
	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return fmt.Errorf("certificate file is not in PEM format")
	}

	if certBlock.Type != "CERTIFICATE" {
		return fmt.Errorf("PEM block is not a certificate (type: %s)", certBlock.Type)
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse PEM certificate: %w", err)
	}

	// Parse the private key - PEM format only
	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return fmt.Errorf("private key file is not in PEM format")
	}

	var privateKey interface{}
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		// PKCS1 format
		key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse PEM RSA private key: %w", err)
		}
		privateKey = key
	case "PRIVATE KEY":
		// PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse PEM PKCS8 private key: %w", err)
		}
		privateKey = key
	default:
		return fmt.Errorf("unsupported PEM block type for private key: %s", keyBlock.Type)
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
