package app

import (
	"fmt"
	"os"
)

// Config holds the application configuration settings.
// It provides a centralised way to manage environment-specific configuration
// with sensible defaults for development and production environments.
type Config struct {
	Host string // Server host address (default: 0.0.0.0 for all interfaces)
	Port string // Server port number (default: 8080)
	Env  string // Environment name (default: development)
}

// LoadConfig loads configuration from environment variables with sensible defaults.
// This function ensures the application has valid configuration even when
// environment variables are not explicitly set.
func LoadConfig() Config {
	return Config{
		Host: getenv("HOST", "0.0.0.0"),    // Bind to all network interfaces
		Port: getenv("PORT", "8080"),       // Standard development port
		Env:  getenv("ENV", "development"), // Default to development mode
	}
}

// Addr returns the formatted address string for the HTTP server.
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
