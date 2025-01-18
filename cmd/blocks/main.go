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
	"com.perkunas/internal/models/block"
	"com.perkunas/internal/models/receipt"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:embed sql/blocks.sql
var blocksSql string

//go:embed genesis.json
var genesisJson string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logger.WithJSONFormat().With(slog.String("scope", "block-svc"))

	db, err := dbConnect(ctx, "blocks.db", blocksSql)
	if err != nil {
		log.Error("failed connecting to blocks.db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	b := &Blocks{
		log:          log,
		blockModel:   block.Model{DB: db},
		receiptModel: receipt.Model{DB: db},
		db:           db,
	}

	if err := b.ensureGenesisBlock(ctx); err != nil {
		b.log.Error("create genesis block fail", "err", err)
		os.Exit(1)
	}

	flag.StringVar(&b.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	b.log.Info("rpc server started", "port exposed", b.apiPort)
	if err := serve(b.apiPort, b); err != nil {
		b.log.Error("failed to start grpc server", "err", err)
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

func serve(port string, service proto.BlockServiceServer) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed starting net listener %w", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)

	proto.RegisterBlockServiceServer(server, service)

	return server.Serve(listener)
}
