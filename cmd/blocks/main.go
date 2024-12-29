package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"com.perkunas/internal/grpcserver"
	"com.perkunas/internal/logger"
	"com.perkunas/internal/models/block"
	"com.perkunas/internal/sqlite"
	"com.perkunas/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:embed sql/blocks.sql
var blocksSql string

//go:embed genesis.json
var genesisJson string

type App struct {
	proto.UnimplementedTransactionServiceServer
	log        *slog.Logger
	blockModel block.BlockModel
	apiPort    string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &App{
		log: logger.WithJSONFormat().With(slog.String("scope", "block-svc")),
	}

	blocksDB, err := dbConnect(ctx, "blocks.db", blocksSql)
	if err != nil {
		app.log.Error("failed connecting to blocks.db", "err", err)
		os.Exit(1)
	}
	defer blocksDB.Close()

	app.blockModel = block.BlockModel{DB: blocksDB}
	if err := app.ensureGenesisBlock(ctx); err != nil {
		app.log.Error("create genesis block fail", "err", err)
		os.Exit(1)
	}

	flag.StringVar(&app.apiPort, "apiport", os.Getenv("API_PORT"), "api port")
	app.log.Info("rpc server started", "port exposed", app.apiPort)
	if err := grpcserver.Serve(app.apiPort, app); err != nil {
		app.log.Error("failed to start grpc server", "err", err)
		os.Exit(1)
	}
}

func (app *App) ensureGenesisBlock(ctx context.Context) error {
	hasGenesis, err := app.blockModel.HasGenesisBlock(ctx)
	if err != nil {
		return fmt.Errorf("unable to check for genesis block presence %w", err)
	}

	if !hasGenesis {
		// create genesis block
		var block block.BlockDB
		if err := json.Unmarshal([]byte(genesisJson), &block); err != nil {
			return fmt.Errorf("unable to unmarshal genesis block json %w", err)
		}

		blockHash, err := block.CalculateHash()
		if err != nil {
			return fmt.Errorf("unable to calculate genesis block hash %w", err)
		}

		block.Hash = blockHash
		if err := app.blockModel.Save(ctx, block); err != nil {
			return fmt.Errorf("unable to persist genesis block %w", err)
		}
	}

	return nil
}

func (app *App) CreateBlock(ctx context.Context, in *proto.CreateBlockRequest) (*proto.CreateBlockResponse, error) {
	txsJson, err := json.Marshal(in.Block.Transactions)
	if err != nil {
		app.log.Info("failed to Marshal in.Block.Transactions", "err", err)
		return nil, status.Error(codes.Internal, "failed processing request payload")
	}

	pld := block.BlockDB{
		Block: block.Block{
			Hash:       in.Block.Hash,
			PrevHash:   in.Block.PrevHash,
			MerkleRoot: in.Block.MerkleRoot,
			Timestamp:  in.Block.Timestamp,
			Height:     in.Block.Height,
			Nonce:      in.Block.Nonce,
		},
		TransactionsDB: txsJson,
	}

	if err := app.blockModel.Save(ctx, pld); err != nil {
		app.log.Info("failed persisting block data", "err", err, "pld", pld)
		return nil, status.Error(codes.Internal, "failed persisting block")
	}

	return &proto.CreateBlockResponse{Hash: pld.Hash, Height: pld.Height}, nil
}

func dbConnect(ctx context.Context, dbName, sql string) (*sqlite.DB, error) {
	db, err := sqlite.NewDB(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to %s db %w", dbName, err)
	}

	if _, err := db.Exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("failed migrating %s db %w", dbName, err)
	}

	return db, nil
}
