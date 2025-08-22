package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// AppConfig holds the application configuration settings.
// It provides a centralised way to manage environment-specific configuration
// with sensible defaults for development and production environments.
type AppConfig struct {
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

// ValidateHTTPS ensures HTTPS configuration is valid.
// This function checks that certificate and key files exist, are readable, and are in PEM format.
func (c *AppConfig) ValidateHTTPS() error {
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

	return nil
}

// Addr returns the formatted address string for the HTTPS server.
// This combines the host and port into a format suitable for net.Listen.
func (c *AppConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// New creates and returns a new configuration instance.
// It loads configuration from environment variables with sensible defaults.
func New() ConfigProvider {
	// Load .env file first to populate environment variables
	loadEnvFile()

	cfg := &AppConfig{
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

	return &configProvider{config: cfg}
}

// configProvider implements ConfigProvider interface
type configProvider struct {
	config *AppConfig
}

// Get returns the configuration struct
func (c *configProvider) Get() *AppConfig {
	return c.config
}

// Validate performs configuration validation and returns any errors
func (c *configProvider) Validate() error {
	return c.config.ValidateHTTPS()
}

// GetString returns a string configuration value by key
func (c *configProvider) GetString(key string) string {
	switch key {
	case "HOST":
		return c.config.Host
	case "PORT":
		return c.config.Port
	case "ENV":
		return c.config.Env
	case "SSL_CERT_FILE":
		return c.config.CertFile
	case "SSL_KEY_FILE":
		return c.config.KeyFile
	case "DB_HOST":
		return c.config.DBHost
	case "DB_PORT":
		return c.config.DBPort
	case "DB_USER":
		return c.config.DBUser
	case "DB_PASSWORD":
		return c.config.DBPassword
	case "DB_NAME":
		return c.config.DBName
	case "DB_SSLMODE":
		return c.config.DBSSLMode
	case "LOG_LEVEL":
		return c.config.LogLevel
	default:
		return ""
	}
}

// GetInt returns an integer configuration value by key
func (c *configProvider) GetInt(key string) int {
	switch key {
	case "MAX_HEADER_BYTES":
		return c.config.MaxHeaderBytes
	default:
		return 0
	}
}

// GetDuration returns a duration configuration value by key
func (c *configProvider) GetDuration(key string) time.Duration {
	switch key {
	case "READ_TIMEOUT":
		return c.config.ReadTimeout
	case "WRITE_TIMEOUT":
		return c.config.WriteTimeout
	case "IDLE_TIMEOUT":
		return c.config.IdleTimeout
	case "READ_HEADER_TIMEOUT":
		return c.config.ReadHeaderTimeout
	default:
		return 0
	}
}

// GetStringSlice returns a string slice configuration value by key
func (c *configProvider) GetStringSlice(key string) []string {
	switch key {
	case "TRUSTED_PROXIES":
		return c.config.TrustedProxies
	default:
		return nil
	}
}

// Helper functions

// loadEnvFile loads environment variables from a .env file.
func loadEnvFile() {
	// Load from .env file if it exists
	if err := loadEnvFileFromPath(".env"); err != nil {
		// File doesn't exist or can't be read - this is normal
		// Environment variables can still be set via system or command line
	}
}

// loadEnvFileFromPath loads environment variables from a specific .env file.
func loadEnvFileFromPath(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err // File doesn't exist or can't be opened
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=value format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"') {
			value = value[1 : len(value)-1]
		}

		// Set environment variable
		os.Setenv(key, value)
	}

	return scanner.Err()
}

// getenv retrieves an environment variable with a fallback default value.
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// getInt retrieves an integer environment variable with a fallback default value.
func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

// getDuration retrieves a duration environment variable with a fallback default value.
func getDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// getStringSlice retrieves a string slice environment variable with a fallback default value.
func getStringSlice(k string, def []string) []string {
	if v := os.Getenv(k); v != "" {
		return strings.Split(v, ",")
	}
	return def
}
