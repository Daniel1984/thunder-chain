package db

import (
	"context"
	"fmt"
	"net/url"
	"runtime"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type DB struct {
	WriteDB *sqlx.DB
	ReadDB  *sqlx.DB
}

func (db *DB) Close() error {
	if err := db.ReadDB.Close(); err != nil {
		return fmt.Errorf("failed closing read db %w", err)
	}

	if err := db.WriteDB.Close(); err != nil {
		return fmt.Errorf("failed closing write db %w", err)
	}

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.ReadDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed pinging reader db %w", err)
	}

	if err := db.WriteDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed pinging writer db %w", err)
	}

	return nil
}

func NewDB(ctx context.Context, dbPath string) (*DB, error) {
	params := url.Values{}
	params.Add("_pragma", "journal_mode=WAL")
	params.Add("_pragma", "busy_timeout=10000")
	params.Add("_pragma", "synchronous=NORMAL")
	params.Add("_pragma", "foreign_keys=ON")

	connectionURL := fmt.Sprintf("file:%s?%s", dbPath, params.Encode())

	writeDB, err := sqlx.Open("sqlite", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create write db: %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	readDB, err := sqlx.Open("sqlite", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create read db: %w", err)
	}
	readDB.SetMaxOpenConns(max(2, runtime.NumCPU()-1))

	db := &DB{
		WriteDB: writeDB,
		ReadDB:  readDB,
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func Connect(ctx context.Context, dbName, sql string) (*DB, error) {
	db, err := NewDB(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to %s db %w", dbName, err)
	}

	if _, err := db.WriteDB.ExecContext(ctx, sql); err != nil {
		return nil, fmt.Errorf("failed migrating %s db %w", dbName, err)
	}

	return db, nil
}
