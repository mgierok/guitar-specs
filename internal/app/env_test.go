package app

import (
	"os"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary .env file for testing
	envContent := `# This is a comment
HOST=127.0.0.1
PORT=9000
ENV=test
SSL_CERT_FILE=/test/cert.crt
SSL_KEY_FILE=/test/key.key

# Empty line should be skipped

# Malformed line should be skipped
MALFORMED_LINE

# Quoted values should work
QUOTED_VALUE="quoted string"
`

	// Write test .env file
	err := os.WriteFile(".env.test", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(".env.test")

	// Clear environment variables that might interfere
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")
	os.Unsetenv("SSL_CERT_FILE")
	os.Unsetenv("SSL_KEY_FILE")
	os.Unsetenv("QUOTED_VALUE")

	// Load the test .env file
	err = loadEnvFile(".env.test")
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}

	// Check that environment variables were set correctly
	tests := map[string]string{
		"HOST":          "127.0.0.1",
		"PORT":          "9000",
		"ENV":           "test",
		"SSL_CERT_FILE": "/test/cert.crt",
		"SSL_KEY_FILE":  "/test/key.key",
		"QUOTED_VALUE":  "quoted string",
	}

	for key, expectedValue := range tests {
		actualValue := os.Getenv(key)
		if actualValue != expectedValue {
			t.Errorf("Environment variable %s: expected '%s', got '%s'", key, expectedValue, actualValue)
		}
	}

	// Check that malformed line was not processed
	if os.Getenv("MALFORMED_LINE") != "" {
		t.Error("Malformed line should not set environment variable")
	}
}

func TestLoadEnvFileNonExistent(t *testing.T) {
	// Try to load a non-existent file
	err := loadEnvFile(".env.nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestLoadEnvFileSimple(t *testing.T) {
	// Create a simple .env file for testing
	envContent := `HOST=127.0.0.1
PORT=9000
ENV=test
`

	// Write test file
	os.WriteFile(".env", []byte(envContent), 0644)
	defer os.Remove(".env")

	// Clear environment completely
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")

	// Load .env file
	LoadEnvFile()

	// Check that environment variables were set correctly
	if host := os.Getenv("HOST"); host != "127.0.0.1" {
		t.Errorf("Expected HOST from .env (127.0.0.1), got '%s'", host)
	}

	if port := os.Getenv("PORT"); port != "9000" {
		t.Errorf("Expected PORT from .env (9000), got '%s'", port)
	}

	if env := os.Getenv("ENV"); env != "test" {
		t.Errorf("Expected ENV from .env (test), got '%s'", env)
	}
}

func TestLoadEnvFileOverwritesExisting(t *testing.T) {
	// Set an environment variable first
	os.Setenv("EXISTING_VAR", "existing_value")

	// Create .env file with the same variable
	envContent := `EXISTING_VAR=new_value
NEW_VAR=new_value
`

	os.WriteFile(".env.test", []byte(envContent), 0644)
	defer os.Remove(".env.test")

	// Load .env file
	loadEnvFile(".env.test")

	// Existing variable should be overwritten (new behavior)
	if value := os.Getenv("EXISTING_VAR"); value != "new_value" {
		t.Errorf("Existing environment variable should be overwritten, got '%s'", value)
	}

	// New variable should be set
	if value := os.Getenv("NEW_VAR"); value != "new_value" {
		t.Errorf("New environment variable should be set, got '%s'", value)
	}
}
