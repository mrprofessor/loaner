package store

import "github.com/jackc/pgx/v5/pgxpool"

// Store provides access to the loan_repayment database.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a Store backed by the given connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
