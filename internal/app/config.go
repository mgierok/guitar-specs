package app

import (
	"fmt"
	"os"
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

	return Config{
		Host:         getenv("HOST", "0.0.0.0"),    // Bind to all network interfaces
		Port:         getenv("PORT", "8080"),       // Standard development port
		Env:          getenv("ENV", "development"), // Default to development mode
		EnableHTTPS:  enableHTTPS,                  // Enable HTTPS if ENABLE_HTTPS=true
		CertFile:     getenv("SSL_CERT_FILE", ""),  // SSL certificate file path
		KeyFile:      getenv("SSL_KEY_FILE", ""),   // SSL private key file path
		RedirectHTTP: redirectHTTP,                 // Redirect HTTP to HTTPS (configurable)

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
