package app

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables
	originalEnv := map[string]string{
		"HOST":          os.Getenv("HOST"),
		"PORT":          os.Getenv("PORT"),
		"ENV":           os.Getenv("ENV"),
		"SSL_CERT_FILE": os.Getenv("SSL_CERT_FILE"),
		"SSL_KEY_FILE":  os.Getenv("SSL_KEY_FILE"),
	}

	// Clean up after test
	defer func() {
		for k, v := range originalEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	t.Run("default values", func(t *testing.T) {
		// Clear all environment variables
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("SSL_CERT_FILE")
		os.Unsetenv("SSL_KEY_FILE")

		cfg := LoadConfig()

		if cfg.Host != "0.0.0.0" {
			t.Errorf("Expected Host '0.0.0.0', got '%s'", cfg.Host)
		}
		if cfg.Port != "8443" {
			t.Errorf("Expected Port '8443', got '%s'", cfg.Port)
		}
		if cfg.Env != "development" {
			t.Errorf("Expected Env 'development', got '%s'", cfg.Env)
		}
		if cfg.CertFile != "" {
			t.Errorf("Expected empty CertFile, got '%s'", cfg.CertFile)
		}
		if cfg.KeyFile != "" {
			t.Errorf("Expected empty KeyFile, got '%s'", cfg.KeyFile)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		os.Setenv("HOST", "127.0.0.1")
		os.Setenv("PORT", "9000")
		os.Setenv("ENV", "production")
		os.Setenv("SSL_CERT_FILE", "/path/to/cert.crt")
		os.Setenv("SSL_KEY_FILE", "/path/to/key.key")

		cfg := LoadConfig()

		if cfg.Host != "127.0.0.1" {
			t.Errorf("Expected Host '127.0.0.1', got '%s'", cfg.Host)
		}
		if cfg.Port != "9000" {
			t.Errorf("Expected Port '9000', got '%s'", cfg.Port)
		}
		if cfg.Env != "production" {
			t.Errorf("Expected Env 'production', got '%s'", cfg.Env)
		}
		if cfg.CertFile != "/path/to/cert.crt" {
			t.Errorf("Expected CertFile '/path/to/cert.crt', got '%s'", cfg.CertFile)
		}
		if cfg.KeyFile != "/path/to/key.key" {
			t.Errorf("Expected KeyFile '/path/to/key.key', got '%s'", cfg.KeyFile)
		}
	})
}

func TestConfig_Addr(t *testing.T) {
	cfg := Config{
		Host: "localhost",
		Port: "8443",
	}

	expected := "localhost:8443"
	if addr := cfg.Addr(); addr != expected {
		t.Errorf("Expected Addr '%s', got '%s'", expected, addr)
	}
}

func TestConfig_ValidateHTTPS(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg := Config{
			CertFile: "/path/to/cert.crt",
			KeyFile:  "/path/to/key.key",
		}

		// This will fail because files don't exist, but we can test the validation logic
		err := cfg.ValidateHTTPS()
		if err == nil {
			t.Error("Expected error for non-existent files, got nil")
		}
		if err.Error() != "SSL certificate file not found: /path/to/cert.crt" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("missing certificate file", func(t *testing.T) {
		cfg := Config{
			CertFile: "",
			KeyFile:  "/path/to/key.key",
		}

		err := cfg.ValidateHTTPS()
		if err == nil {
			t.Error("Expected error for missing certificate file, got nil")
		}
		if err.Error() != "SSL_CERT_FILE not specified" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("missing private key file", func(t *testing.T) {
		cfg := Config{
			CertFile: "/path/to/cert.crt",
			KeyFile:  "",
		}

		err := cfg.ValidateHTTPS()
		if err == nil {
			t.Error("Expected error for missing private key file, got nil")
		}
		if err.Error() != "SSL_KEY_FILE not specified" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}
