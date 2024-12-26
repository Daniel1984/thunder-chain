package main

import (
	"context"
	"fmt"
	"log/slog"

	"com.perkunas/internal/models"
	"com.perkunas/internal/sqlite"
	"com.perkunas/proto"
)

type Mempool struct {
	proto.UnimplementedTransactionServiceServer
	log     *slog.Logger
	db      *sqlite.DB
	models  *models.Models
	apiPort string
}

func (mp *Mempool) CreateTransaction(ctx context.Context, in *proto.CreateTransactionRequest) (*proto.CreateTransactionResponse, error) {
	mp.log.Info("::: CreateTransaction...")
	fmt.Printf("::: CreateTransaction... in :> %+v\n", in.Transaction)
	return &proto.CreateTransactionResponse{Hash: "foobarbazqux"}, nil
}

func (mp *Mempool) DeleteTransaction(ctx context.Context, in *proto.DeleteTransactionRequest) (*proto.DeleteTransactionResponse, error) {
	mp.log.Info("::: DeleteTransaction...")
	return &proto.DeleteTransactionResponse{Success: true}, nil
}
