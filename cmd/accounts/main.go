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
	"com.perkunas/internal/models/account"
	"com.perkunas/internal/models/balancechange"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:embed sql/accounts.sql
var accountsSql string

const dbName = "accounts.db"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.WithJSONFormat().With(slog.String("scope", "accounts-svc"))

	db, err := dbConnect(ctx, dbName, accountsSql)
	if err != nil {
		log.Error(fmt.Sprintf("failed connecting to %s", dbName), "err", err)
		os.Exit(1)
	}
	defer db.Close()

	app := &App{
		db:                 db,
		log:                log,
		accModel:           &account.Model{DB: db},
		balanceChangeModel: &balancechange.Model{DB: db},
	}

	flag.StringVar(&app.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	app.log.Info("rpc server started", "port exposed", app.apiPort)
	if err := serve(app.apiPort, app); err != nil {
		app.log.Error("failed to start grpc server", "err", err)
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

func serve(port string, service proto.BalanceChangeServiceServer) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed starting net listener %w", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)

	proto.RegisterBalanceChangeServiceServer(server, service)

	return server.Serve(listener)
}
