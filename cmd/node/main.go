package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"com.perkunas/internal/logger"
	"com.perkunas/internal/middleware"
	"com.perkunas/internal/server"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	log        *slog.Logger
	mempoolAPI string
	apiPort    string
	rpcClient  proto.MempoolServiceClient
}

func main() {
	app := &App{log: logger.WithJSONFormat().With(slog.String("scope", "node"))}
	flag.StringVar(&app.mempoolAPI, "mempoolapi", os.Getenv("MEMPOOL_API"), "mempool api endpoint")
	flag.StringVar(&app.apiPort, "apiport", os.Getenv("API_PORT"), "node api port")

	conn, client, err := rpcClient(app.mempoolAPI)
	if err != nil {
		app.log.Error("grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer conn.Close()
	app.rpcClient = client

	srv := httpServer(app.getRouter(), app.apiPort)
	app.log.Info("api server started", "port exposed", app.apiPort)
	if err := srv.Start(); err != nil {
		app.log.Error("failed starting server", "err", err)
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

func rpcClient(apiUrl string) (*grpc.ClientConn, proto.MempoolServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewMempoolServiceClient(conn)
	return conn, client, nil
}
