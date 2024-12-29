package main

import (
	"context"
	_ "embed"
	"flag"
	"log/slog"
	"os"

	"com.perkunas/internal/grpcserver"
	"com.perkunas/internal/logger"
	"com.perkunas/internal/models"
	"com.perkunas/internal/sqlite"
)

//go:embed sql/mempool.sql
var mempoolsql string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.WithJSONFormat().With(slog.String("scope", "mempool"))

	// initialize sqlite db for mempool persistance layer
	db, err := sqlite.NewDB(ctx, "mempool.db")
	if err != nil {
		log.Error("failed connecting DB", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// run migrations
	if _, err := db.Exec(ctx, mempoolsql); err != nil {
		log.Error("failed migrating db", "err", err)
		os.Exit(1)
	}

	mempoolSvc := &Mempool{
		log:    log,
		db:     db,
		models: models.NewModels(db),
	}

	flag.StringVar(&mempoolSvc.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	log.Info("rpc server started", "port exposed", mempoolSvc.apiPort)
	if err := grpcserver.Serve(mempoolSvc.apiPort, mempoolSvc); err != nil {
		log.Error("failed to start grpc server", "err", err)
		os.Exit(1)
	}
}
