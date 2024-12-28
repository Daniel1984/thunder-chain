package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"com.perkunas/internal/logger"
	"com.perkunas/internal/models"
	"com.perkunas/internal/sqlite"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", mempoolSvc.apiPort))
	if err != nil {
		log.Error("failed starting net listener", "err", err)
		os.Exit(1)
	}

	server := grpc.NewServer()
	reflection.Register(server)

	log.Info("rpc server started", "port exposed", mempoolSvc.apiPort)
	proto.RegisterTransactionServiceServer(server, mempoolSvc)
	if err := server.Serve(listener); err != nil {
		log.Error("failed to serve", "err", err)
		os.Exit(1)
	}
}
