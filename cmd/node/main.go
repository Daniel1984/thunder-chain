package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"com.perkunas/internal/logger"
	"com.perkunas/internal/middleware"
	"com.perkunas/internal/server"
	"com.perkunas/internal/sqlite"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed sql/accounts.sql
var accountsSql string

type App struct {
	log        *slog.Logger
	mempoolAPI string
	apiPort    string
	rpcClient  proto.TransactionServiceClient
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.WithJSONFormat().With(slog.String("scope", "node"))

	accountsDB, err := dbConnection(ctx, "accounts.db", accountsSql)
	if err != nil {
		log.Error("failed connecting to db", "err", err)
		os.Exit(1)
	}
	defer accountsDB.Close()

	app := &App{log: log}

	flag.StringVar(&app.mempoolAPI, "mempoolapi", os.Getenv("MEMPOOL_API"), "mempool api endpoint")
	flag.StringVar(&app.apiPort, "apiport", os.Getenv("API_PORT"), "node api port")

	conn, client, err := rpcClient(app.mempoolAPI)
	if err != nil {
		log.Error("grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer conn.Close()
	app.rpcClient = client

	srv := httpServer(app.getRouter(), app.apiPort)

	log.Info("api server started", "port exposed", app.apiPort)
	if err := srv.Start(); err != nil {
		log.Error("failed starting server", "err", err)
		os.Exit(1)
	}
}

func httpServer(mux *http.ServeMux, port string) *server.Server {
	return server.
		Get().
		WithAddr(fmt.Sprintf(":%s", port)).
		WithMiddleware(middleware.Chain(middleware.LogReq)).
		WithRouter(mux)
}

func rpcClient(apiUrl string) (*grpc.ClientConn, proto.TransactionServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewTransactionServiceClient(conn)
	return conn, client, nil
}

func dbConnection(ctx context.Context, dbName, sql string) (*sqlite.DB, error) {
	db, err := sqlite.NewDB(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to %s db %w", dbName, err)
	}

	if _, err := db.Exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("failed migrating %s db %w", dbName, err)
	}

	return db, nil
}
