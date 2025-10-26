package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"com.perkunas/internal/db"
	"com.perkunas/internal/logger"
	"com.perkunas/internal/models/account"
	"com.perkunas/internal/models/balancechange"
	"com.perkunas/internal/models/block"
	"com.perkunas/internal/models/genesisblock"
	"com.perkunas/internal/models/receipt"
)

//go:embed sql/state.sql
var stateSql string

//go:embed genesis.json
var genesisJson string

var dbPath string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.WithJSONFormat().With(slog.String("scope", "state-svc"))

	flag.StringVar(&dbPath, "db-path", os.Getenv("DB_PATH"), "state db absolute path")
	db, err := dbConnect(ctx, dbPath, stateSql)
	if err != nil {
		log.Error(fmt.Sprintf("failed connecting to %s", dbPath), "err", err)
		os.Exit(1)
	}
	defer db.Close()

	s := &State{
		db:                 db,
		log:                log,
		accModel:           &account.Model{DB: db},
		balanceChangeModel: &balancechange.Model{DB: db},
		blockModel:         &block.Model{DB: db},
		genesisBlockModel:  &genesisblock.Model{DB: db},
		receiptModel:       &receipt.Model{DB: db},
	}

	if err := s.ensureGenesisBlock(ctx); err != nil {
		s.log.Error("failed to create genesis block", "err", err)
		os.Exit(1)
	}

	flag.StringVar(&s.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	if err := s.Start(); err != nil {
		log.Error("failed to start grpc server", "err", err)
		os.Exit(1)
	}
}

func dbConnect(ctx context.Context, dbName, sql string) (*db.DB, error) {
	db, err := db.NewDB(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to %s db %w", dbName, err)
	}

	if _, err := db.WriteDB.ExecContext(ctx, sql); err != nil {
		return nil, fmt.Errorf("failed migrating %s db %w", dbName, err)
	}

	return db, nil
}
