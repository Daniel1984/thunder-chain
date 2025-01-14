package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"com.perkunas/internal/db"
	"com.perkunas/internal/models/block"
	"com.perkunas/internal/models/receipt"
	"com.perkunas/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	proto.UnimplementedBlockServiceServer
	log          *slog.Logger
	blockModel   block.Model
	receiptModel receipt.Model
	db           *db.DB
	apiPort      string
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
		block.TransactionsDB = "[]"
		if err := app.blockModel.Save(ctx, block); err != nil {
			return fmt.Errorf("unable to persist genesis block %w", err)
		}
	}

	return nil
}

func (app *App) CreateBlock(ctx context.Context, in *proto.CreateBlockRequest) (*proto.CreateBlockResponse, error) {
	b := in.GetBlock()
	if b == nil {
		app.log.Info("missing block data")
		return nil, status.Error(codes.Aborted, "missing block data")
	}

	txs := b.GetTransactions()
	if len(txs) == 0 {
		app.log.Info("missing block transactions")
		return nil, status.Error(codes.Aborted, "missing block transactions")
	}

	txsJson, err := json.Marshal(txs)
	if err != nil {
		app.log.Info("failed to Marshal txs", "err", err)
		return nil, status.Error(codes.Internal, "failed to Marshal txs")
	}

	blockPld := block.BlockDB{
		Block: block.Block{
			Hash:       b.GetHash(),
			PrevHash:   b.GetPrevHash(),
			MerkleRoot: b.GetMerkleRoot(),
			Timestamp:  b.GetTimestamp(),
			Height:     b.GetHeight(),
			Nonce:      b.GetNonce(),
		},
		TransactionsDB: string(txsJson),
	}

	dbTx, err := app.db.WriteDB.BeginTxx(ctx, nil)
	if err != nil {
		app.log.Error("failed to begin DB transaction", "err", err)
		return nil, status.Error(codes.Internal, "failed to begin DB transaction")
	}

	if err := app.blockModel.SaveWithTX(ctx, dbTx, blockPld); err != nil {
		dbTx.Rollback()
		app.log.Info("failed persisting block data", "err", err, "height", blockPld.Height, "hash", blockPld.Hash)
		return nil, status.Error(codes.Internal, "failed persisting block")
	}

	receiptsPld := receipt.ProtoToReceipts(txs, b.GetHash())
	if err := app.receiptModel.InsertBatch(ctx, dbTx, receiptsPld); err != nil {
		dbTx.Rollback()
		app.log.Info("failed persisting receipts", "err", err)
		return nil, status.Error(codes.Internal, "failed persisting receipts")
	}

	if err := dbTx.Commit(); err != nil {
		app.log.Error("failed storing block and receipts data", "err", err)
		return nil, status.Error(codes.Internal, "failed storing block and receipts data")
	}

	return &proto.CreateBlockResponse{Hash: blockPld.Hash, Height: blockPld.Height}, nil
}

func (app *App) GetLatestBlock(ctx context.Context, in *proto.GetLatestBlockRequest) (*proto.GetLatestBlockResponse, error) {
	block, err := app.blockModel.GetLatest(ctx)
	if err != nil {
		app.log.Info("failed getting latest block data", "err", err)
		return nil, status.Error(codes.Internal, "failed getting latest block data")
	}

	return &proto.GetLatestBlockResponse{
		Block: &proto.Block{
			Hash:       block.Hash,
			PrevHash:   block.PrevHash,
			MerkleRoot: block.MerkleRoot,
			Height:     block.Height,
			Nonce:      block.Nonce,
			Timestamp:  block.Timestamp,
		},
	}, nil
}
