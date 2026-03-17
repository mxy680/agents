package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the interface that both pgx.Conn and pgxpool.Pool satisfy.
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// Queries provides all database query methods.
type Queries struct {
	db DBTX
}

// New creates a new Queries instance.
func New(db DBTX) *Queries {
	return &Queries{db: db}
}
