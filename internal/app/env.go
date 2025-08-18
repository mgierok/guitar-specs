package app

import (
	"bufio"
	"os"
	"strings"
)

// loadEnvFile loads environment variables from a .env file.
// This function reads the .env file line by line and sets environment variables
// for the current process. It supports standard .env format with KEY=value pairs.
func loadEnvFile(filename string) error {
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

// LoadEnvFile loads environment variables from a single .env file.
// This simplifies configuration management by using only one environment file.
func LoadEnvFile() {
	// Load from .env file if it exists
	if err := loadEnvFile(".env"); err != nil {
		// File doesn't exist or can't be read - this is normal
		// Environment variables can still be set via system or command line
	}
}
