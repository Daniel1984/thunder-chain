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
	proto.UnimplementedBalanceChangeServiceServer
	log                *slog.Logger
	apiPort            string
	db                 *db.DB
	accModel           *account.Model
	balanceChangeModel *balancechange.Model
}

func (app *App) CreateBalanceChange(ctx context.Context, in *proto.CreateBalanceChangeRequest) (*proto.CreateBalanceChangeResponse, error) {
	bc := in.GetBalancechange()
	if bc == nil {
		app.log.Info("missing balance change data")
		return nil, status.Error(codes.Aborted, "missing balance change data")
	}

	txs := bc.GetTransactions()
	if txs == nil {
		app.log.Info("balance change missing transactions")
		return nil, status.Error(codes.Aborted, "balance change missing transactions")
	}

	sort.Slice(txs, func(i, j int) bool {
		return txs[i].GetTimestamp() < txs[j].GetTimestamp()
	})

	dbTx, err := app.db.WriteDB.BeginTxx(ctx, nil)
	if err != nil {
		app.log.Info("failed to begin DB transaction", "err", err)
		return nil, status.Error(codes.Internal, "failed to begin DB transaction")
	}
	defer dbTx.Rollback()

	for _, tx := range txs {
		fromAcc, err := app.accModel.UpsertNoUpdate(ctx, dbTx, tx.GetFromAddr())
		if err != nil {
			app.log.Info("failed to get src account", "err", err)
			return nil, status.Error(codes.Internal, "failed to get src account")
		}

		toAcc, err := app.accModel.UpsertNoUpdate(ctx, dbTx, tx.GetToAddr())
		if err != nil {
			app.log.Info("failed to get dest account", "err", err)
			return nil, status.Error(codes.Internal, "failed to get dest account")
		}

		fromAccBc := balancechange.BalanceChange{
			PreviousBalance: fromAcc.Balance,
			NewBalance:      fromAcc.Balance - tx.GetAmount() - tx.GetFee(),
			ChangeAmount:    -(tx.GetAmount() + tx.GetFee()),
			AccountID:       fromAcc.ID,
			BlockHeight:     bc.GetBlockHeight(),
			BlockHash:       bc.GetBlockHash(),
			TxHash:          tx.GetHash(),
			Timestamp:       tx.GetTimestamp(),
		}

		if err := app.balanceChangeModel.Crete(ctx, dbTx, fromAccBc); err != nil {
			app.log.Info("failed to create source acc balance change record", "err", err)
			return nil, status.Error(codes.Internal, "failed to create source acc balance change record")
		}

		if _, err := app.accModel.Upsert(ctx, dbTx, account.Account{
			Address: fromAcc.Address,
			Balance: fromAccBc.NewBalance,
			Nonce:   fromAcc.Nonce + 1,
		}); err != nil {
			app.log.Info("failed to update source account balance", "err", err)
			return nil, status.Error(codes.Internal, "failed to update source account balance")
		}

		toAccBc := balancechange.BalanceChange{
			PreviousBalance: toAcc.Balance,
			NewBalance:      toAcc.Balance + tx.GetAmount(),
			ChangeAmount:    tx.GetAmount(),
			AccountID:       toAcc.ID,
			BlockHeight:     bc.GetBlockHeight(),
			BlockHash:       bc.GetBlockHash(),
			TxHash:          tx.GetHash(),
			Timestamp:       tx.GetTimestamp(),
		}

		if err := app.balanceChangeModel.Crete(ctx, dbTx, toAccBc); err != nil {
			app.log.Info("failed to create destination acc balance change record", "err", err)
			return nil, status.Error(codes.Internal, "failed to create destination acc balance change record")
		}

		if _, err := app.accModel.Upsert(ctx, dbTx, account.Account{
			Address: toAcc.Address,
			Balance: toAccBc.NewBalance,
			Nonce:   toAcc.Nonce + 1,
		}); err != nil {
			app.log.Info("failed to update destination account balance", "err", err)
			return nil, status.Error(codes.Internal, "failed to update destination account balance")
		}
	}

	return &proto.CreateBalanceChangeResponse{Success: true, Message: "STATE_UPDATED"}, nil
}
