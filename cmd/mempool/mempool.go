package main

import (
	"context"
	"log/slog"

	"com.perkunas/internal/models/transaction"
	"com.perkunas/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Mempool struct {
	proto.UnimplementedTransactionServiceServer
	log     *slog.Logger
	txModel transaction.Model
	apiPort string
}

func (mp *Mempool) CreateTransaction(ctx context.Context, in *proto.CreateTransactionRequest) (*proto.CreateTransactionResponse, error) {
	pld := transaction.Transaction{
		Hash:      in.Transaction.Hash,
		From:      in.Transaction.FromAddr,
		To:        in.Transaction.ToAddr,
		Signature: in.Transaction.Signature,
		Amount:    in.Transaction.Amount,
		Fee:       in.Transaction.Fee,
		Nonce:     in.Transaction.Nonce,
		Data:      in.Transaction.Data,
		Timestamp: in.Transaction.Timestamp,
		Expires:   in.Transaction.Expires,
	}

	if err := mp.txModel.Save(ctx, pld); err != nil {
		mp.log.Info("failed saving transaction in mempool", "err", err)
		return nil, status.Error(codes.Internal, "failed persisting transaction")
	}

	return &proto.CreateTransactionResponse{Hash: pld.Hash}, nil
}

// TODO: implement and accept list of tx hashes to delete
func (mp *Mempool) DeleteTransaction(ctx context.Context, in *proto.DeleteTransactionRequest) (*proto.DeleteTransactionResponse, error) {
	return &proto.DeleteTransactionResponse{Success: true}, nil
}
