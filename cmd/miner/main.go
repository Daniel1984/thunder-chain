package main

import (
	"context"
	"log/slog"
	"time"

	"com.perkunas/internal/logger"
	"com.perkunas/internal/models/block"
	"com.perkunas/proto"
)

type App struct {
	log        *slog.Logger
	mempoolAPI string
	blockModel block.Model
	rpcClient  proto.TransactionServiceClient
}

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.WithJSONFormat().With(slog.String("scope", "node"))

	// blocksDB, err := dbConnection(ctx, "blocks.db", blocksSql)
	// if err != nil {
	// 	log.Error("failed connecting to db", "err", err)
	// 	os.Exit(1)
	// }
	// defer blocksDB.Close()
	for {
		time.Sleep(10 * time.Second)
		log.Info("slept for 10s")
	}
}
