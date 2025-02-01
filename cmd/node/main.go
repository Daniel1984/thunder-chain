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
	log              *slog.Logger
	apiPort          string
	mempoolAPI       string
	blocksAPI        string
	peerNodes        []peernode.Node
	mempoolRpcClient proto.MempoolServiceClient
	blocksRPC        proto.BlockServiceClient
}

func main() {
	n := &Node{log: logger.WithJSONFormat().With(slog.String("scope", "node"))}
	flag.StringVar(&n.mempoolAPI, "mempoolapi", os.Getenv("MEMPOOL_API"), "mempool api endpoint")
	flag.StringVar(&n.apiPort, "apiport", os.Getenv("API_PORT"), "node api port")
	flag.StringVar(&n.blocksAPI, "blocksapi", os.Getenv("BLOCKS_API"), "blocks api endpoint")

	memPoolConn, client, err := mempoolRpcClient(n.mempoolAPI)
	if err != nil {
		n.log.Error("grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer memPoolConn.Close()
	n.mempoolRpcClient = client

	blocksConn, blocksClient, err := blocksRPCClient(n.blocksAPI)
	if err != nil {
		n.log.Error("blocks grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer blocksConn.Close()
	n.blocksRPC = blocksClient

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

func blocksRPCClient(apiUrl string) (*grpc.ClientConn, proto.BlockServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	cli := proto.NewBlockServiceClient(conn)
	return conn, cli, nil
}
