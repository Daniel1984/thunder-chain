package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"com.perkunas/internal/models/block"
	"com.perkunas/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	proto.UnimplementedBlockServiceServer
	log        *slog.Logger
	blockModel block.Model
	apiPort    string
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
