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
	"com.perkunas/internal/middleware"
	"com.perkunas/internal/models/peernode"
	"com.perkunas/internal/server"
	"github.com/ethereum/go-ethereum/log"
	_ "modernc.org/sqlite"
)

var (
	//go:embed sql/bootstrapnode.sql
	bootstrapnodesql string
	dbPath           string
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bn := &BootstrapNode{
		log: logger.WithJSONFormat().With(slog.String("scope", "bootstrapnode")),
	}

	flag.StringVar(&dbPath, "db-path", os.Getenv("DB_PATH"), "bootstrapnode db absolute path")
	flag.StringVar(&bn.port, "port", os.Getenv("PORT"), "bootstrap node port")
	flag.Parse()

	bootstrapDb, err := db.Connect(ctx, dbPath, bootstrapnodesql)
	if err != nil {
		log.Error(fmt.Sprintf("failed connecting to %s", dbPath), "err", err)
		os.Exit(1)
	}
	defer bootstrapDb.Close()

	bn.peerModel = &peernode.Model{DB: bootstrapDb}

	go bn.cleanupInactivePeers(ctx)

	srv := server.
		Get().
		WithAddr(fmt.Sprintf(":%s", bn.port)).
		WithMiddleware(middleware.Chain(middleware.LogReq)).
		WithRouter(bn.getRouter())

	bn.log.Info("bootstrap node starting...", "port", bn.port)
	if err := srv.Start(); err != nil {
		bn.log.Error("failed starting bootstrap server", "err", err)
		os.Exit(1)
	}
}
