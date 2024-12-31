package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"com.perkunas/internal/db"
	"com.perkunas/internal/logger"
	"com.perkunas/internal/models/transaction"
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
	db, err := db.NewDB(ctx, "mempool.db")
	if err != nil {
		log.Error("failed connecting DB", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// run migrations
	if _, err := db.WriteDB.ExecContext(ctx, mempoolsql); err != nil {
		log.Error("failed migrating db", "err", err)
		os.Exit(1)
	}

	mempoolSvc := &Mempool{
		log:     log,
		txModel: transaction.Model{DB: db},
	}

	flag.StringVar(&mempoolSvc.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	log.Info("rpc server started", "port exposed", mempoolSvc.apiPort)
	if err := serve(mempoolSvc.apiPort, mempoolSvc); err != nil {
		log.Error("failed to start grpc server", "err", err)
		os.Exit(1)
	}
}

func serve(port string, service proto.TransactionServiceServer) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed starting net listener %w", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)

	proto.RegisterTransactionServiceServer(server, service)

	return server.Serve(listener)
}
