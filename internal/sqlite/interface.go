package sqlite

import (
	"context"
	"database/sql"
)

// Database defines the interface for database operations
type Database interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	Select(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Close() error
}
