package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseProvider defines the interface for database operations.
// This interface allows for dependency inversion and easier testing.
type DatabaseProvider interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error
	
	// Close closes the database connection
	Close()
	
	// GetPool returns the underlying connection pool
	GetPool() *pgxpool.Pool
	
	// Ping tests the database connection
	Ping(ctx context.Context) error
	
	// IsConnected returns true if the database is connected
	IsConnected() bool
	
	// GetConnectionInfo returns database connection information
	GetConnectionInfo() ConnectionInfo
}

// ConnectionInfo holds database connection information
type ConnectionInfo struct {
	Host     string
	Port     string
	Database string
	User     string
	SSLMode  string
	Connected bool
	ConnectedAt *time.Time
}
