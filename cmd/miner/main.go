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

type App struct {
	log        *slog.Logger
	mempoolAPI string
	stateAPI   string
	blocksAPI  string
	mempoolRPC proto.MempoolServiceClient
	stateRPC   proto.StateChangeServiceClient
	blocksRPC  proto.BlockServiceClient
}

func main() {
	app := &App{log: logger.WithJSONFormat().With(slog.String("scope", "miner-svc"))}
	flag.StringVar(&app.mempoolAPI, "mempoolapi", os.Getenv("MEMPOOL_API"), "mempool api endpoint")
	flag.StringVar(&app.stateAPI, "stateapi", os.Getenv("STATE_API"), "state api endpoint")
	flag.StringVar(&app.blocksAPI, "blocksapi", os.Getenv("BLOCKS_API"), "blocks api endpoint")

	// initiate mempool rpc client
	mempoolConn, mempoolClient, err := mempoolRPCClient(app.mempoolAPI)
	if err != nil {
		app.log.Error("mempool grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer mempoolConn.Close()
	app.mempoolRPC = mempoolClient

	// initiate state rpc client
	stateConn, stateClient, err := stateRPCClient(app.stateAPI)
	if err != nil {
		app.log.Error("state grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer stateConn.Close()
	app.stateRPC = stateClient

	// initiate blocks rpc client
	blocksConn, blocksClient, err := blocksRPCClient(app.blocksAPI)
	if err != nil {
		app.log.Error("blocks grpc did not connect", "err", err)
		os.Exit(1)
	}
	defer blocksConn.Close()
	app.blocksRPC = blocksClient

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		app.log.Error("failed to start the miner", "err", err)
		os.Exit(1)
	}
}

func mempoolRPCClient(apiUrl string) (*grpc.ClientConn, proto.MempoolServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewMempoolServiceClient(conn)
	return conn, client, nil
}

func stateRPCClient(apiUrl string) (*grpc.ClientConn, proto.StateChangeServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewStateChangeServiceClient(conn)
	return conn, client, nil
}

func blocksRPCClient(apiUrl string) (*grpc.ClientConn, proto.BlockServiceClient, error) {
	conn, err := grpc.NewClient(apiUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewBlockServiceClient(conn)
	return conn, client, nil
}
