package store

import "github.com/jackc/pgx/v4/pgxpool"

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
