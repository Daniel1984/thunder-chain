package main

import (
	"context"
	"log/slog"
	"sort"

	"com.perkunas/internal/db"
	"com.perkunas/internal/models/account"
	"com.perkunas/internal/models/balancechange"
	"com.perkunas/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	proto.UnimplementedBlockServiceServer
	log     *slog.Logger
	apiPort string
	txModel account.Model
	db      *db.DB
}

func (app *App) CreateBlock(ctx context.Context, in *proto.CreateBlockRequest) (*proto.CreateBlockResponse, error) {
	block := in.GetBlock()
	if block == nil {
		app.log.Info("missing block data")
		return nil, status.Error(codes.Aborted, "missing block data")
	}

	txs := block.GetTransactions()
	if txs == nil {
		app.log.Info("block missing transactions")
		return nil, status.Error(codes.Aborted, "block missing transactions")
	}

	sort.Slice(txs, func(i, j int) bool {
		return txs[i].GetTimestamp() < txs[j].GetTimestamp()
	})

	for _, tx := range block.Transactions {
		// create new transaction type, maybe accountantTransaction with block info
		// use regular db transaction
		// app.db.WriteDB.BeginTxx()
		txs = append(txs, balancechange.BalanceChange{
			PreviousBalance: 0,
			NewBalance:      0,
			ChangeAmount:    0,
			BlockHeight:     block.GetHeight(),
			BlockHash:       block.GetHash(),
			TxHash:          tx.GetHash(),
			Timestamp:       tx.GetTimestamp(),
		})
	}

	// pld := account.Account{
	// 	Address:   in.Account.Address,
	// 	Balance:   in.Account.Balance,
	// 	Nonce:     in.Account.Nonce,
	// 	Timestamp: in.Account.Timestamp,
	// }

	// if err := app.txModel.Save(ctx, pld); err != nil {
	// 	app.log.Info("failed saving account", "err", err)
	// 	return nil, status.Error(codes.Internal, "failed persisting account")
	// }

	return &proto.CreateBlockResponse{}, nil
}
