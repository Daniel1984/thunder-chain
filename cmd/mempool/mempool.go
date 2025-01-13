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
	proto.UnimplementedMempoolServiceServer
	log     *slog.Logger
	txModel transaction.Model
	apiPort string
}

func (mp *Mempool) CreateMempool(ctx context.Context, in *proto.CreateMempoolRequest) (*proto.CreateMempoolResponse, error) {
	tx := in.GetTransaction()
	if tx == nil {
		mp.log.Error("request payload missing transaction")
		return nil, status.Error(codes.Canceled, "request payload missing transaction")
	}

	pld := transaction.Transaction{
		Hash:      tx.GetHash(),
		From:      tx.GetFromAddr(),
		To:        tx.GetToAddr(),
		Signature: tx.GetSignature(),
		Amount:    tx.GetAmount(),
		Fee:       tx.GetFee(),
		Nonce:     tx.GetNonce(),
		Data:      tx.GetData(),
		Timestamp: tx.GetTimestamp(),
		Expires:   tx.GetExpires(),
	}

	if err := mp.txModel.Save(ctx, pld); err != nil {
		mp.log.Error("failed saving transaction in mempool", "err", err)
		return nil, status.Error(codes.Internal, "failed persisting transaction")
	}

	return &proto.CreateMempoolResponse{Hash: pld.Hash}, nil
}

// TODO: implement and accept list of tx hashes to delete
func (mp *Mempool) DeleteTransaction(ctx context.Context, in *proto.DeleteMempoolRequest) (*proto.DeleteMempoolResponse, error) {
	return &proto.DeleteMempoolResponse{Success: true}, nil
}

func (mp *Mempool) PendingTransactions(ctx context.Context, in *proto.PendingTransactionsRequest) (*proto.PendingTransactionsResponse, error) {
	txs, err := mp.txModel.Pending(ctx)
	if err != nil {
		mp.log.Error("failed getting pending transactions", "err", err)
		return nil, status.Error(codes.Internal, "failed getting pending transactions")
	}

	protoTxs := []*proto.Transaction{}
	for _, tx := range txs {
		protoTxs = append(protoTxs, &proto.Transaction{
			Hash:      tx.Hash,
			FromAddr:  tx.From,
			ToAddr:    tx.To,
			Signature: tx.Signature,
			Fee:       tx.Fee,
			Amount:    tx.Amount,
			Nonce:     tx.Nonce,
			Timestamp: tx.Timestamp,
			Expires:   tx.Expires,
		})
	}

	return &proto.PendingTransactionsResponse{Transactions: protoTxs}, nil
}
