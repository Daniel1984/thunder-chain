package main

import (
	"context"
	"log/slog"

	"com.perkunas/internal/models"
	"com.perkunas/internal/sqlite"
	"com.perkunas/proto"
)

type Mempool struct {
	proto.UnimplementedTransactionServiceServer
	log    *slog.Logger
	db     *sqlite.DB
	models *models.Models
}

func (mp *Mempool) CreateTransaction(ctx context.Context, in *proto.CreateTransactionRequest) (*proto.CreateTransactionResponse, error) {
	mp.log.Info("::: CreateTransaction...")
	return &proto.CreateTransactionResponse{Id: "foobarbazqux"}, nil
}

func (mp *Mempool) DeleteTransaction(ctx context.Context, in *proto.DeleteTransactionRequest) (*proto.DeleteTransactionResponse, error) {
	mp.log.Info("::: DeleteTransaction...")
	return &proto.DeleteTransactionResponse{Success: true}, nil
}
