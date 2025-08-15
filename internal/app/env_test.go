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
ENABLE_HTTPS=true
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
	os.Unsetenv("ENABLE_HTTPS")
	os.Unsetenv("SSL_CERT_FILE")
	os.Unsetenv("SSL_KEY_FILE")
	os.Unsetenv("QUOTED_VALUE")

	// Load the test .env file
	err = loadEnvFile(".env.test", false)
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}

	// Check that environment variables were set correctly
	tests := map[string]string{
		"HOST":          "127.0.0.1",
		"PORT":          "9000",
		"ENV":           "test",
		"ENABLE_HTTPS":  "true",
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
	err := loadEnvFile(".env.nonexistent", false)
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestLoadEnvFilesPriority(t *testing.T) {
	// Create test .env files with different priorities
	envDefault := `HOST=0.0.0.0
PORT=8080
ENV=development
`

	envDev := `HOST=0.0.0.0
PORT=8443
ENV=development
ENABLE_HTTPS=true
`

	envLocal := `HOST=127.0.0.1
PORT=9000
ENV=local
`

	// Write test files
	os.WriteFile(".env", []byte(envDefault), 0644)
	os.WriteFile(".env.development", []byte(envDev), 0644)
	os.WriteFile(".env.local", []byte(envLocal), 0644)

	defer func() {
		os.Remove(".env")
		os.Remove(".env.development")
		os.Remove(".env.local")
	}()

	// Clear environment completely
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")
	os.Unsetenv("ENABLE_HTTPS")

	// Don't set ENV in environment - let .env files set it

	// Load .env files
	LoadEnvFiles()

	// With the new policy, .env.local does NOT override values set by .env/.env.[ENV]
	if host := os.Getenv("HOST"); host != "0.0.0.0" {
		t.Errorf("Expected HOST from .env (0.0.0.0), got '%s'", host)
	}

	if port := os.Getenv("PORT"); port != "8080" {
		t.Errorf("Expected PORT from .env (8080), got '%s'", port)
	}

	if env := os.Getenv("ENV"); env != "development" {
		t.Errorf("Expected ENV from .env (development), got '%s'", env)
	}

	// Check that ENABLE_HTTPS was loaded from .env.development
	// (since ENV=development from .env, it loads .env.development)
	if https := os.Getenv("ENABLE_HTTPS"); https != "true" {
		t.Errorf("Expected ENABLE_HTTPS from .env.development (true), got '%s'", https)
	}
}

func TestLoadEnvFileExistingEnvVar(t *testing.T) {
	// Set an environment variable first
	os.Setenv("EXISTING_VAR", "existing_value")

	// Create .env file with the same variable
	envContent := `EXISTING_VAR=new_value
NEW_VAR=new_value
`

	os.WriteFile(".env.test", []byte(envContent), 0644)
	defer os.Remove(".env.test")

	// Load .env file
	loadEnvFile(".env.test", false)

	// Existing variable should not be overwritten
	if value := os.Getenv("EXISTING_VAR"); value != "existing_value" {
		t.Errorf("Existing environment variable should not be overwritten, got '%s'", value)
	}

	// New variable should be set
	if value := os.Getenv("NEW_VAR"); value != "new_value" {
		t.Errorf("New environment variable should be set, got '%s'", value)
	}
}
