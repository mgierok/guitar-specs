package db

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database represents a database connection manager.
// It implements the DatabaseProvider interface.
type Database struct {
	config      DatabaseConfig
	pool        *pgxpool.Pool
	connected   bool
	connectedAt *time.Time
}

// DatabaseConfig holds database configuration parameters.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// New creates a new database instance with the given configuration.
// It returns a DatabaseProvider interface for dependency injection.
func New(config DatabaseConfig) DatabaseProvider {
	return &Database{
		config: config,
	}
}

// Connect establishes a connection to the database.
// It creates a connection pool and validates the connection.
func (d *Database) Connect(ctx context.Context) error {
	dsn := d.buildDSN()
	if dsn == "" {
		return fmt.Errorf("database configuration missing; set DB_HOST, DB_USER, DB_NAME")
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.pool = pool
	d.connected = true
	now := time.Now()
	d.connectedAt = &now

	return nil
}

// Close closes the database connection and releases resources.
func (d *Database) Close() {
	if d.pool != nil {
		d.pool.Close()
		d.pool = nil
	}
	d.connected = false
	d.connectedAt = nil
}

// GetPool returns the underlying connection pool.
// This method provides access to the pool for direct database operations.
func (d *Database) GetPool() *pgxpool.Pool {
	return d.pool
}

// Ping tests the database connection.
// It returns an error if the connection is not available.
func (d *Database) Ping(ctx context.Context) error {
	if d.pool == nil {
		return fmt.Errorf("database not connected")
	}
	return d.pool.Ping(ctx)
}

// IsConnected returns true if the database is connected.
func (d *Database) IsConnected() bool {
	return d.connected && d.pool != nil
}

// GetConnectionInfo returns database connection information.
// This method provides metadata about the current connection.
func (d *Database) GetConnectionInfo() ConnectionInfo {
	return ConnectionInfo{
		Host:        d.config.Host,
		Port:        d.config.Port,
		Database:    d.config.Database,
		User:        d.config.User,
		SSLMode:     d.config.SSLMode,
		Connected:   d.connected,
		ConnectedAt: d.connectedAt,
	}
}

// buildDSN assembles a PostgreSQL DSN from configuration parameters.
// It returns an empty string if required parameters are missing.
func (d *Database) buildDSN() string {
	if d.config.Host == "" || d.config.User == "" || d.config.Database == "" {
		return ""
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.config.User, d.config.Password),
		Host:   fmt.Sprintf("%s:%s", d.config.Host, d.config.Port),
		Path:   "/" + d.config.Database,
	}

	q := url.Values{}
	if d.config.SSLMode != "" {
		q.Set("sslmode", d.config.SSLMode)
	}

	u.RawQuery = q.Encode()
	return u.String()
}
