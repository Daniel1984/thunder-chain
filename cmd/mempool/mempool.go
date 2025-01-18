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

	pld := transaction.FromProtoTx(tx)
	if err := mp.txModel.Save(ctx, pld); err != nil {
		mp.log.Error("failed saving transaction in mempool", "err", err)
		return nil, status.Error(codes.Internal, "failed persisting transaction")
	}

	return &proto.CreateMempoolResponse{Hash: pld.Hash}, nil
}

func (mp *Mempool) DeleteMempoolBatch(ctx context.Context, in *proto.DeleteMempoolBatchRequest) (*proto.DeleteMempoolBatchResponse, error) {
	if err := mp.txModel.DeleteBatch(ctx, in.Ids); err != nil {
		mp.log.Error("failed deleting batch of transactions", "err", err, "txIDs", in.Ids)
		return nil, status.Error(codes.Internal, "failed deleting batch of transactions")
	}

	return &proto.DeleteMempoolBatchResponse{Success: true, DeletedCount: int32(len(in.Ids))}, nil
}

func (mp *Mempool) PendingTransactions(ctx context.Context, in *proto.PendingTransactionsRequest) (*proto.PendingTransactionsResponse, error) {
	txs, err := mp.txModel.Pending(ctx)
	if err != nil {
		mp.log.Error("failed getting pending transactions", "err", err)
		return nil, status.Error(codes.Internal, "failed getting pending transactions")
	}

	protoTxs := transaction.ToProtoTxs(txs)
	return &proto.PendingTransactionsResponse{Transactions: protoTxs}, nil
}
