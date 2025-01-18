package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"com.perkunas/internal/logger"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	m := &Miner{log: logger.WithJSONFormat().With(slog.String("scope", "miner-svc"))}
	flag.StringVar(&m.mempoolAPI, "mempoolapi", os.Getenv("MEMPOOL_API"), "mempool api endpoint")
	flag.StringVar(&m.stateAPI, "stateapi", os.Getenv("STATE_API"), "state api endpoint")
	flag.StringVar(&m.blocksAPI, "blocksapi", os.Getenv("BLOCKS_API"), "blocks api endpoint")

	// initiate mempool rpc client
	mempoolConn, mempoolClient, err := mempoolRPCClient(m.mempoolAPI)
	if err != nil {
		m.log.Error("mempool grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer mempoolConn.Close()
	m.mempoolRPC = mempoolClient

	// initiate state rpc client
	stateConn, stateClient, err := stateRPCClient(m.stateAPI)
	if err != nil {
		m.log.Error("state grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer stateConn.Close()
	m.stateRPC = stateClient

	// initiate blocks rpc client
	blocksConn, blocksClient, err := blocksRPCClient(m.blocksAPI)
	if err != nil {
		m.log.Error("blocks grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer blocksConn.Close()
	m.blocksRPC = blocksClient

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		m.log.Error("failed to start the miner", "err", err)
		os.Exit(1)
	}
}

func mempoolRPCClient(apiUrl string) (*grpc.ClientConn, proto.MempoolServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	cli := proto.NewMempoolServiceClient(conn)
	return conn, cli, nil
}

func stateRPCClient(apiUrl string) (*grpc.ClientConn, proto.StateServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	cli := proto.NewStateServiceClient(conn)
	return conn, cli, nil
}

func blocksRPCClient(apiUrl string) (*grpc.ClientConn, proto.BlockServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	cli := proto.NewBlockServiceClient(conn)
	return conn, cli, nil
}
