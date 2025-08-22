package config

import (
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
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

		cfg := New()

		if cfg.GetString("HOST") != "0.0.0.0" {
			t.Errorf("Expected Host '0.0.0.0', got '%s'", cfg.GetString("HOST"))
		}
		if cfg.GetString("PORT") != "8443" {
			t.Errorf("Expected Port '8443', got '%s'", cfg.GetString("PORT"))
		}
		if cfg.GetString("ENV") != "development" {
			t.Errorf("Expected Env 'development', got '%s'", cfg.GetString("ENV"))
		}
		if cfg.GetString("SSL_CERT_FILE") != "" {
			t.Errorf("Expected empty SSL_CERT_FILE, got '%s'", cfg.GetString("SSL_CERT_FILE"))
		}
		if cfg.GetString("SSL_KEY_FILE") != "" {
			t.Errorf("Expected empty SSL_KEY_FILE, got '%s'", cfg.GetString("SSL_KEY_FILE"))
		}
	})

	t.Run("custom values", func(t *testing.T) {
		os.Setenv("HOST", "127.0.0.1")
		os.Setenv("PORT", "9000")
		os.Setenv("ENV", "production")
		os.Setenv("SSL_CERT_FILE", "/path/to/cert.crt")
		os.Setenv("SSL_KEY_FILE", "/path/to/key.key")

		cfg := New()

		if cfg.GetString("HOST") != "127.0.0.1" {
			t.Errorf("Expected Host '127.0.0.1', got '%s'", cfg.GetString("HOST"))
		}
		if cfg.GetString("PORT") != "9000" {
			t.Errorf("Expected Port '9000', got '%s'", cfg.GetString("PORT"))
		}
		if cfg.GetString("ENV") != "production" {
			t.Errorf("Expected Env 'production', got '%s'", cfg.GetString("ENV"))
		}
		if cfg.GetString("SSL_CERT_FILE") != "/path/to/cert.crt" {
			t.Errorf("Expected SSL_CERT_FILE '/path/to/cert.crt', got '%s'", cfg.GetString("SSL_CERT_FILE"))
		}
		if cfg.GetString("SSL_KEY_FILE") != "/path/to/key.key" {
			t.Errorf("Expected SSL_KEY_FILE '/path/to/key.key', got '%s'", cfg.GetString("SSL_KEY_FILE"))
		}
	})
}

func TestConfigProvider_Get(t *testing.T) {
	cfg := New()
	appConfig := cfg.Get()

	if appConfig == nil {
		t.Error("Expected non-nil AppConfig, got nil")
	}

	if appConfig.Host == "" {
		t.Error("Expected non-empty Host, got empty")
	}
}

func TestConfigProvider_GetDuration(t *testing.T) {
	cfg := New()

	readTimeout := cfg.GetDuration("READ_TIMEOUT")
	if readTimeout != 10*time.Second {
		t.Errorf("Expected READ_TIMEOUT 10s, got %v", readTimeout)
	}

	writeTimeout := cfg.GetDuration("WRITE_TIMEOUT")
	if writeTimeout != 30*time.Second {
		t.Errorf("Expected WRITE_TIMEOUT 30s, got %v", writeTimeout)
	}
}

func TestConfigProvider_GetInt(t *testing.T) {
	cfg := New()

	maxHeaderBytes := cfg.GetInt("MAX_HEADER_BYTES")
	if maxHeaderBytes != 1<<20 {
		t.Errorf("Expected MAX_HEADER_BYTES 1MB, got %d", maxHeaderBytes)
	}
}

func TestConfigProvider_GetStringSlice(t *testing.T) {
	cfg := New()

	trustedProxies := cfg.GetStringSlice("TRUSTED_PROXIES")
	if len(trustedProxies) != 2 {
		t.Errorf("Expected 2 trusted proxies, got %d", len(trustedProxies))
	}

	expected := []string{"127.0.0.1", "::1"}
	for i, proxy := range expected {
		if trustedProxies[i] != proxy {
			t.Errorf("Expected trusted proxy %s at index %d, got %s", proxy, i, trustedProxies[i])
		}
	}
}
