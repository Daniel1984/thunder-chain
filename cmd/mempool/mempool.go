package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"com.perkunas/internal/models/transaction"
	"com.perkunas/internal/scheduler"
	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
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

func (mp *Mempool) SpawnCleanupJob(ctx context.Context) *scheduler.Job {
	cleanupJob := &scheduler.Job{
		Interval: time.Minute,
		Task: func(ctx context.Context) {
			if res, err := mp.txModel.ClearExpired(ctx); err != nil {
				mp.log.Error("failed clearing expired transactions", "err", err)
			} else {
				affected, _ := res.RowsAffected()
				if affected > 0 {
					mp.log.Info("cleared expired transactions", "count", affected)
				}
			}
		},
	}

	cleanupJob.Start(ctx)
	return cleanupJob
}

func (mp *Mempool) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", mp.apiPort))
	if err != nil {
		return fmt.Errorf("failed starting net listener %w", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	proto.RegisterMempoolServiceServer(server, mp)

	mp.log.Info("rpc server started", "port exposed", mp.apiPort)
	return server.Serve(listener)
}
