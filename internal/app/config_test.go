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
		"ENABLE_HTTPS":  os.Getenv("ENABLE_HTTPS"),
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
		os.Unsetenv("ENABLE_HTTPS")
		os.Unsetenv("SSL_CERT_FILE")
		os.Unsetenv("SSL_KEY_FILE")

		cfg := LoadConfig()

		if cfg.Host != "0.0.0.0" {
			t.Errorf("Expected Host '0.0.0.0', got '%s'", cfg.Host)
		}
		if cfg.Port != "8080" {
			t.Errorf("Expected Port '8080', got '%s'", cfg.Port)
		}
		if cfg.Env != "development" {
			t.Errorf("Expected Env 'development', got '%s'", cfg.Env)
		}
		if cfg.EnableHTTPS != false {
			t.Errorf("Expected EnableHTTPS false, got %v", cfg.EnableHTTPS)
		}
		if cfg.CertFile != "" {
			t.Errorf("Expected empty CertFile, got '%s'", cfg.CertFile)
		}
		if cfg.KeyFile != "" {
			t.Errorf("Expected empty KeyFile, got '%s'", cfg.KeyFile)
		}
		if cfg.RedirectHTTP != true {
			t.Errorf("Expected RedirectHTTP true (default), got %v", cfg.RedirectHTTP)
		}
	})

	t.Run("HTTPS enabled", func(t *testing.T) {
		os.Setenv("ENABLE_HTTPS", "true")
		os.Setenv("SSL_CERT_FILE", "/path/to/cert.crt")
		os.Setenv("SSL_KEY_FILE", "/path/to/key.key")

		cfg := LoadConfig()

		if cfg.EnableHTTPS != true {
			t.Errorf("Expected EnableHTTPS true, got %v", cfg.EnableHTTPS)
		}
		if cfg.CertFile != "/path/to/cert.crt" {
			t.Errorf("Expected CertFile '/path/to/cert.crt', got '%s'", cfg.CertFile)
		}
		if cfg.KeyFile != "/path/to/key.key" {
			t.Errorf("Expected KeyFile '/path/to/key.key', got '%s'", cfg.KeyFile)
		}
		if cfg.RedirectHTTP != true {
			t.Errorf("Expected RedirectHTTP true, got %v", cfg.RedirectHTTP)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		os.Setenv("HOST", "127.0.0.1")
		os.Setenv("PORT", "9000")
		os.Setenv("ENV", "production")

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
	})
}

func TestConfig_Addr(t *testing.T) {
	cfg := Config{
		Host: "localhost",
		Port: "8080",
	}

	expected := "localhost:8080"
	if addr := cfg.Addr(); addr != expected {
		t.Errorf("Expected Addr '%s', got '%s'", expected, addr)
	}
}

func TestConfig_AddrHTTPS(t *testing.T) {
	cfg := Config{
		Host: "0.0.0.0",
		Port: "8443",
	}

	expected := "0.0.0.0:8443"
	if addr := cfg.AddrHTTPS(); addr != expected {
		t.Errorf("Expected AddrHTTPS '%s', got '%s'", expected, addr)
	}
}

func TestConfig_AddrHTTP(t *testing.T) {
	cfg := Config{
		Host: "example.com",
		Port: "8443", // This should be ignored for HTTP addr
	}

	expected := "example.com:80"
	if addr := cfg.AddrHTTP(); addr != expected {
		t.Errorf("Expected AddrHTTP '%s', got '%s'", expected, addr)
	}
}

func TestConfig_ValidateHTTPS(t *testing.T) {
	t.Run("HTTPS disabled should pass validation", func(t *testing.T) {
		cfg := Config{
			EnableHTTPS: false,
		}

		if err := cfg.ValidateHTTPS(); err != nil {
			t.Errorf("Expected no error for disabled HTTPS, got %v", err)
		}
	})

	t.Run("HTTPS enabled without cert should fail", func(t *testing.T) {
		cfg := Config{
			EnableHTTPS: true,
			CertFile:    "",
			KeyFile:     "",
		}

		if err := cfg.ValidateHTTPS(); err == nil {
			t.Error("Expected error for missing certificate file")
		}
	})

	t.Run("HTTPS enabled without key should fail", func(t *testing.T) {
		cfg := Config{
			EnableHTTPS: true,
			CertFile:    "/path/to/cert.crt",
			KeyFile:     "",
		}

		if err := cfg.ValidateHTTPS(); err == nil {
			t.Error("Expected error for missing key file")
		}
	})

	t.Run("HTTPS enabled with non-existent files should fail", func(t *testing.T) {
		cfg := Config{
			EnableHTTPS: true,
			CertFile:    "/nonexistent/cert.crt",
			KeyFile:     "/nonexistent/key.key",
		}

		if err := cfg.ValidateHTTPS(); err == nil {
			t.Error("Expected error for non-existent files")
		}
	})
}
