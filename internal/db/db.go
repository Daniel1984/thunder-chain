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
	connectionUrlParams := make(url.Values)
	connectionUrlParams.Add("_txlock", "immediate")
	connectionUrlParams.Add("_journal_mode", "WAL")
	connectionUrlParams.Add("_busy_timeout", "5000")
	connectionUrlParams.Add("_synchronous", "NORMAL")
	connectionUrlParams.Add("_cache_size", "1000000000")
	connectionUrlParams.Add("_foreign_keys", "true")
	connectionUrl := fmt.Sprintf("file:%s?%s", dbPath, connectionUrlParams.Encode())

	writeDB, err := sqlx.Open("sqlite", connectionUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to establish write db connection %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	readDB, err := sqlx.Open("sqlite", connectionUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to establish read db connection %w", err)
	}
	readDB.SetMaxOpenConns(max(4, runtime.NumCPU()))

	db := &DB{
		WriteDB: writeDB,
		ReadDB:  readDB,
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
