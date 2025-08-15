package app

import (
	"bufio"
	"os"
	"strings"
)

// loadEnvFile loads environment variables from a .env file.
// This function reads the .env file line by line and sets environment variables
// for the current process. It supports standard .env format with KEY=value pairs.
// The force parameter determines whether to overwrite existing environment variables.
func loadEnvFile(filename string, force bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return err // File doesn't exist or can't be opened
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
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

		// Set environment variable (overwrite if force=true)
		if force || os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// LoadEnvFiles attempts to load environment variables from .env files.
// Load order and override policy:
// 1. .env (base defaults)
// 2. .env.[ENVIRONMENT] (environment-specific defaults)
// 3. .env.local (developer-local additions; does not override existing values)
func LoadEnvFiles() {
	// Try to load default .env file first to get basic configuration
	_ = loadEnvFile(".env", false)

	// Try to load environment-specific .env file
	env := os.Getenv("ENV")
	if env == "" {
		env = "development" // Default environment
	}
	_ = loadEnvFile(".env."+env, false)

	// Load .env.local last, but do not override previously set values
	_ = loadEnvFile(".env.local", false)
}
