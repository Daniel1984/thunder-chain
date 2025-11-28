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
	"com.perkunas/internal/models/transaction"
)

var (
	//go:embed sql/mempool.sql
	mempoolsql string
	dbPath     string
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.WithJSONFormat().With(slog.String("scope", "mempool"))
	flag.StringVar(&dbPath, "db-path", os.Getenv("DB_PATH"), "mempool db absolute path")

	mempoolDb, err := db.Connect(ctx, dbPath, mempoolsql)
	if err != nil {
		log.Error(fmt.Sprintf("failed connecting to %s", dbPath), "err", err)
		os.Exit(1)
	}
	defer mempoolDb.Close()

	mempoolSvc := &Mempool{
		log:     log,
		txModel: transaction.Model{DB: mempoolDb},
	}

	cleanupJob := mempoolSvc.SpawnCleanupJob(ctx)
	defer cleanupJob.Stop()

	flag.StringVar(&mempoolSvc.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	if err := mempoolSvc.Start(); err != nil {
		log.Error("failed to start grpc server", "err", err)
		os.Exit(1)
	}
}
