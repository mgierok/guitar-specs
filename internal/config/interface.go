package config

import "time"

// ConfigProvider defines the interface for application configuration.
// This interface allows for dependency inversion and easier testing.
type ConfigProvider interface {
	// Get returns the configuration struct
	Get() *AppConfig

	// Validate performs configuration validation and returns any errors
	Validate() error

	// GetString returns a string configuration value by key
	GetString(key string) string

	// GetInt returns an integer configuration value by key
	GetInt(key string) int

	// GetDuration returns a duration configuration value by key
	GetDuration(key string) time.Duration

	// GetStringSlice returns a string slice configuration value by key
	GetStringSlice(key string) []string
}
