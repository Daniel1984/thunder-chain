package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"com.perkunas/internal/logger"
	"com.perkunas/internal/middleware"
	"com.perkunas/internal/models/peernode"
	"com.perkunas/internal/server"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	n := &Node{
		log:       logger.WithJSONFormat().With(slog.String("scope", "node")),
		peerNodes: []peernode.Peer{},
	}

	flag.StringVar(&n.mempoolAPI, "mempoolapi", os.Getenv("MEMPOOL_API"), "mempool api endpoint")
	flag.StringVar(&n.stateAPI, "stateapi", os.Getenv("STATE_API"), "state api endpoint")
	flag.StringVar(&n.bootstrapAPI, "bootstrapapi", os.Getenv("BOOTSTRAP_API"), "bootstrap node api endpoint")
	flag.StringVar(&n.apiPort, "apiport", os.Getenv("API_PORT"), "node api port")
	flag.BoolVar(&n.participateInPeerDiscovery, "peer-discovery", os.Getenv("PEER_DISCOVERY") == "true", "participate in peer discovery")
	flag.Parse()

	_, mempoolClient, err := mempoolRpcClient(n.mempoolAPI)
	if err != nil {
		n.log.Error("mempool grpc connection failed", "err", err)
		os.Exit(1)
	}
	n.mempoolRPC = mempoolClient

	_, stateClient, err := stateRPCClient(n.stateAPI)
	if err != nil {
		n.log.Error("state grpc connection failed", "err", err)
		os.Exit(1)
	}
	n.stateRPC = stateClient

	// get IP upon starting program to avoid multiple lookups later
	locIP, err := GetOutboundIP()
	if err != nil {
		n.log.Error("failed to get outbound ip", "err", err)
		os.Exit(1)
	}
	n.outboundIP = locIP

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if n.participateInPeerDiscovery {
		n.RegisterAsPeer(ctx)
		go n.StartHeartbeat(ctx)
	}

	n.FetchPeers(ctx)
	go n.StartPeerRefresher(ctx)

	srv := server.
		Get().
		WithAddr(fmt.Sprintf(":%s", n.apiPort)).
		WithMiddleware(middleware.Chain(middleware.LogReq)).
		WithRouter(n.getRouter())

	n.log.Info("node starting", "port", n.apiPort)
	if err := srv.Start(); err != nil {
		n.log.Error("failed starting server", "err", err)
		os.Exit(1)
	}
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
