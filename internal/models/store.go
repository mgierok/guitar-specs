package models

import "github.com/jackc/pgx/v5/pgxpool"

// Store aggregates all model stores backed by a shared pgx connection pool.
type Store struct {
	DB      *pgxpool.Pool
	Guitars GuitarStore
}

// NewStore constructs a Store with initialised repositories.
func NewStore(db *pgxpool.Pool) *Store {
	s := &Store{DB: db}
	s.Guitars = GuitarStore{DB: db}
	return s
}
