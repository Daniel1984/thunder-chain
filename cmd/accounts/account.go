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
	proto.UnimplementedStateChangeServiceServer
	log                *slog.Logger
	apiPort            string
	db                 *db.DB
	accModel           *account.Model
	balanceChangeModel *balancechange.Model
}

func (app *App) GetAccountByAddress(ctx context.Context, in *proto.GetAccountByAddressRequest) (*proto.GetAccountByAddressResponse, error) {
	acc, err := app.accModel.Get(ctx, in.GetAddress())
	if err != nil {
		app.log.Error("failed to get account", "err", err, "addr", in.GetAddress())
		return nil, status.Error(codes.Internal, "failed to get account")
	}

	return &proto.GetAccountByAddressResponse{
		Account: acc.ToProto(),
	}, nil
}

func (app *App) CreateStateChange(ctx context.Context, in *proto.CreateStateChangeRequest) (*proto.CreateStateChangeResponse, error) {
	sc := in.GetStatechange()
	if sc == nil {
		app.log.Info("missing balance change data")
		return nil, status.Error(codes.Aborted, "missing balance change data")
	}

	txs := sc.GetTransactions()
	if len(txs) == 0 {
		app.log.Info("balance change missing transactions")
		return nil, status.Error(codes.Aborted, "balance change missing transactions")
	}

	sort.Slice(txs, func(i, j int) bool {
		return txs[i].GetTimestamp() < txs[j].GetTimestamp()
	})

	dbTx, err := app.db.WriteDB.BeginTxx(ctx, nil)
	if err != nil {
		app.log.Error("failed to begin DB transaction", "err", err)
		return nil, status.Error(codes.Internal, "failed to begin DB transaction")
	}

	for _, tx := range txs {
		fromAcc, err := app.accModel.UpsertNoUpdate(ctx, dbTx, tx.GetFromAddr())
		if err != nil {
			app.log.Error("failed to get src account", "err", err)
			dbTx.Rollback()
			return nil, status.Error(codes.Internal, "failed to get src account")
		}

		toAcc, err := app.accModel.UpsertNoUpdate(ctx, dbTx, tx.GetToAddr())
		if err != nil {
			app.log.Error("failed to get dest account", "err", err)
			dbTx.Rollback()
			return nil, status.Error(codes.Internal, "failed to get dest account")
		}

		fromAccBc := balancechange.BalanceChange{
			PreviousBalance: fromAcc.Balance,
			NewBalance:      fromAcc.Balance - tx.GetAmount() - tx.GetFee(),
			ChangeAmount:    -(tx.GetAmount() + tx.GetFee()),
			AccountID:       fromAcc.ID,
			BlockHeight:     sc.GetBlockHeight(),
			BlockHash:       sc.GetBlockHash(),
			TxHash:          tx.GetHash(),
			Timestamp:       tx.GetTimestamp(),
		}

		if err := app.balanceChangeModel.Crete(ctx, dbTx, fromAccBc); err != nil {
			app.log.Error("failed to create source acc balance change record", "err", err)
			dbTx.Rollback()
			return nil, status.Error(codes.Internal, "failed to create source acc balance change record")
		}

		if _, err := app.accModel.Upsert(ctx, dbTx, account.Account{
			Address: fromAcc.Address,
			Balance: fromAccBc.NewBalance,
			Nonce:   fromAcc.Nonce + 1,
		}); err != nil {
			app.log.Error("failed to update source account balance", "err", err)
			dbTx.Rollback()
			return nil, status.Error(codes.Internal, "failed to update source account balance")
		}

		toAccBc := balancechange.BalanceChange{
			PreviousBalance: toAcc.Balance,
			NewBalance:      toAcc.Balance + tx.GetAmount(),
			ChangeAmount:    tx.GetAmount(),
			AccountID:       toAcc.ID,
			BlockHeight:     sc.GetBlockHeight(),
			BlockHash:       sc.GetBlockHash(),
			TxHash:          tx.GetHash(),
			Timestamp:       tx.GetTimestamp(),
		}

		if err := app.balanceChangeModel.Crete(ctx, dbTx, toAccBc); err != nil {
			app.log.Error("failed to create destination acc balance change record", "err", err)
			dbTx.Rollback()
			return nil, status.Error(codes.Internal, "failed to create destination acc balance change record")
		}

		if _, err := app.accModel.Upsert(ctx, dbTx, account.Account{
			Address: toAcc.Address,
			Balance: toAccBc.NewBalance,
			Nonce:   toAcc.Nonce + 1,
		}); err != nil {
			app.log.Error("failed to update destination account balance", "err", err)
			dbTx.Rollback()
			return nil, status.Error(codes.Internal, "failed to update destination account balance")
		}
	}

	if err := dbTx.Commit(); err != nil {
		app.log.Error("failed updating block state", "err", err)
		return nil, status.Error(codes.Internal, "failed updating block state")
	}

	return &proto.CreateStateChangeResponse{Message: "STATE_UPDATED"}, nil
}
