package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"com.perkunas/internal/grpcserver"
	"com.perkunas/internal/logger"
	"com.perkunas/internal/models/block"
	"com.perkunas/internal/sqlite"
)

//go:embed sql/blocks.sql
var blocksSql string

//go:embed genesis.json
var genesisJson string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &App{
		log: logger.WithJSONFormat().With(slog.String("scope", "block-svc")),
	}

	blocksDB, err := dbConnect(ctx, "blocks.db", blocksSql)
	if err != nil {
		app.log.Error("failed connecting to blocks.db", "err", err)
		os.Exit(1)
	}
	defer blocksDB.Close()

	app.blockModel = block.BlockModel{DB: blocksDB}
	if err := app.ensureGenesisBlock(ctx); err != nil {
		app.log.Error("create genesis block fail", "err", err)
		os.Exit(1)
	}

	flag.StringVar(&app.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	app.log.Info("rpc server started", "port exposed", app.apiPort)
	if err := grpcserver.Serve(app.apiPort, app); err != nil {
		app.log.Error("failed to start grpc server", "err", err)
		os.Exit(1)
	}
}

func dbConnect(ctx context.Context, dbName, sql string) (*sqlite.DB, error) {
	db, err := sqlite.NewDB(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to %s db %w", dbName, err)
	}

	if _, err := db.Exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("failed migrating %s db %w", dbName, err)
	}

	return db, nil
}
