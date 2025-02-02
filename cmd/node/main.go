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
	"com.perkunas/internal/models/peernode"
	"com.perkunas/internal/server"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	proto.UnimplementedNodeServiceServer
	log        *slog.Logger
	apiPort    string
	mempoolAPI string
	stateAPI   string
	peerNodes  []peernode.Node
	mempoolRPC proto.MempoolServiceClient
	stateRPC   proto.StateServiceClient
}

func main() {
	n := &Node{log: logger.WithJSONFormat().With(slog.String("scope", "node"))}
	flag.StringVar(&n.mempoolAPI, "mempoolapi", os.Getenv("MEMPOOL_API"), "mempool api endpoint")
	flag.StringVar(&n.stateAPI, "stateapi", os.Getenv("STATE_API"), "state api endpoint")
	flag.StringVar(&n.apiPort, "apiport", os.Getenv("API_PORT"), "node api port")

	// initiate mempool rpc client
	memPoolConn, client, err := mempoolRpcClient(n.mempoolAPI)
	if err != nil {
		n.log.Error("grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer memPoolConn.Close()
	n.mempoolRPC = client

	// initiate state rpc client
	stateConn, stateClient, err := stateRPCClient(n.stateAPI)
	if err != nil {
		n.log.Error("state grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer stateConn.Close()
	n.stateRPC = stateClient

	srv := httpServer(n.getRouter(), n.apiPort)
	n.log.Info("api server started", "port exposed", n.apiPort)
	if err := srv.Start(); err != nil {
		n.log.Error("failed starting server", "err", err)
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

func mempoolRpcClient(apiUrl string) (*grpc.ClientConn, proto.MempoolServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewMempoolServiceClient(conn)
	return conn, client, nil
}

func stateRPCClient(apiUrl string) (*grpc.ClientConn, proto.StateServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	cli := proto.NewStateServiceClient(conn)
	return conn, cli, nil
}
