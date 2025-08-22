package db

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	db := New(config)

	if db == nil {
		t.Error("Expected non-nil DatabaseProvider, got nil")
	}

	// Test that it implements the interface
	var _ DatabaseProvider = db
}

func TestDatabase_Connect_InvalidConfig(t *testing.T) {
	// Test with missing required fields
	config := DatabaseConfig{
		Host:     "", // Missing host
		User:     "testuser",
		Database: "testdb",
	}

	db := New(config)
	ctx := context.Background()

	err := db.Connect(ctx)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	if db.IsConnected() {
		t.Error("Expected database to not be connected")
	}
}

func TestDatabase_ConnectionInfo(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	db := New(config)
	info := db.GetConnectionInfo()

	// Test connection info before connection
	if info.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", info.Host)
	}
	if info.Port != "5432" {
		t.Errorf("Expected port '5432', got '%s'", info.Port)
	}
	if info.Database != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", info.Database)
	}
	if info.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", info.User)
	}
	if info.SSLMode != "disable" {
		t.Errorf("Expected SSL mode 'disable', got '%s'", info.SSLMode)
	}
	if info.Connected {
		t.Error("Expected not connected before Connect()")
	}
	if info.ConnectedAt != nil {
		t.Error("Expected ConnectedAt to be nil before connection")
	}
}

func TestDatabase_IsConnected(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	db := New(config)

	// Should not be connected initially
	if db.IsConnected() {
		t.Error("Expected database to not be connected initially")
	}

	// Test with invalid config (should fail to connect)
	ctx := context.Background()
	err := db.Connect(ctx)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	// Should still not be connected after failed connection attempt
	if db.IsConnected() {
		t.Error("Expected database to not be connected after failed connection")
	}
}

func TestDatabase_Close(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	db := New(config)

	// Close should not panic even if not connected
	db.Close()

	// Should not be connected after close
	if db.IsConnected() {
		t.Error("Expected database to not be connected after Close()")
	}
}

func TestDatabase_GetPool(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	db := New(config)

	// GetPool should return nil before connection
	pool := db.GetPool()
	if pool != nil {
		t.Error("Expected nil pool before connection")
	}
}

func TestDatabase_Ping(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	db := New(config)
	ctx := context.Background()

	// Ping should fail before connection
	err := db.Ping(ctx)
	if err == nil {
		t.Error("Expected error when pinging before connection")
	}
}

func TestDatabase_BuildDSN(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	db := &Database{config: config}
	
	// Use reflection to access private method for testing
	// This is a bit of a hack, but allows us to test the DSN building logic
	dsn := db.buildDSN()
	
	expectedContains := []string{
		"postgres://",
		"testuser:testpass@",
		"localhost:5432",
		"/testdb",
		"sslmode=disable",
	}
	
	for _, expected := range expectedContains {
		if !contains(dsn, expected) {
			t.Errorf("Expected DSN to contain '%s', got '%s'", expected, dsn)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
