package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sort"

	"com.perkunas/internal/db"
	"com.perkunas/internal/models/account"
	"com.perkunas/internal/models/balancechange"
	"com.perkunas/internal/models/block"
	"com.perkunas/internal/models/genesisblock"
	"com.perkunas/internal/models/receipt"
	"com.perkunas/proto"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type State struct {
	proto.UnimplementedStateServiceServer
	log                *slog.Logger
	apiPort            string
	db                 *db.DB
	accModel           *account.Model
	blockModel         *block.Model
	genesisBlockModel  *genesisblock.Model
	receiptModel       *receipt.Model
	balanceChangeModel *balancechange.Model
}

func (s *State) ensureGenesisBlock(ctx context.Context) error {
	hasGenesis, err := s.genesisBlockModel.HasGenesisBlock(ctx)
	if err != nil {
		return fmt.Errorf("unable to check for genesis block presence %w", err)
	}

	if !hasGenesis {
		// create genesis block
		var gBlock genesisblock.GenesisBlock
		if err := json.Unmarshal([]byte(genesisJson), &gBlock); err != nil {
			return fmt.Errorf("unable to unmarshal genesis block json %w", err)
		}

		blockHash, err := gBlock.CalculateHash()
		if err != nil {
			return fmt.Errorf("unable to calculate genesis block hash %w", err)
		}

		gBlock.Hash = blockHash
		gBlock.TransactionsDB = "[]"

		dbTx, err := s.db.WriteDB.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin DB transaction %w", err)
		}

		if err := s.genesisBlockModel.SaveWithTX(ctx, dbTx, gBlock); err != nil {
			dbTx.Rollback()
			return fmt.Errorf("unable to persist genesis block %w", err)
		}

		if err := s.accModel.BatchInsert(ctx, dbTx, gBlock.Accounts); err != nil {
			dbTx.Rollback()
			return fmt.Errorf("unable to persist genesis accounts %w", err)
		}

		if err := dbTx.Commit(); err != nil {
			return fmt.Errorf("failed creating genesis block %w", err)
		}
	}

	return nil
}

func (s *State) GetAccountByAddress(ctx context.Context, in *proto.AccountByAddressReq) (*proto.AccountByAddressRes, error) {
	acc, err := s.accModel.Get(ctx, in.GetAddress())
	if err != nil {
		s.log.Error("failed to get account", "err", err, "addr", in.GetAddress())
		return nil, status.Error(codes.Internal, "failed to get account")
	}

	return &proto.AccountByAddressRes{
		Account: acc.ToProto(),
	}, nil
}

func (s *State) CreateBlock(ctx context.Context, in *proto.CreateBlockReq) (*proto.CreateBlockRes, error) {
	block := in.GetBlock()
	if block == nil {
		s.log.Info("missing balance change data")
		return &proto.CreateBlockRes{Message: "NO_STATE_DATA"}, nil
	}

	txs := block.GetTransactions()
	if len(txs) == 0 {
		s.log.Info("missing transactions to update balances")
		return &proto.CreateBlockRes{Message: "MISSING_STATE_TXS"}, nil
	}

	// order by timestamp so we process older txs first
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].GetTimestamp() < txs[j].GetTimestamp()
	})

	dbTx, err := s.db.WriteDB.BeginTxx(ctx, nil)
	if err != nil {
		s.log.Error("failed to begin DB transaction", "err", err)
		return nil, status.Error(codes.Internal, "failed to begin DB transaction")
	}

	if err := s.updateBalances(ctx, dbTx, txs, block); err != nil {
		s.log.Error("failed updating balances", "err", err)
		dbTx.Rollback()
		return nil, status.Error(codes.Internal, "failed updating balances")
	}

	if err := s.createBlock(ctx, dbTx, txs, block); err != nil {
		s.log.Error("failed creating block", "err", err)
		dbTx.Rollback()
		return nil, status.Error(codes.Internal, "failed creating block")
	}

	if err := dbTx.Commit(); err != nil {
		s.log.Error("failed creating block", "err", err)
		return nil, status.Error(codes.Internal, "failed creating block")
	}

	return &proto.CreateBlockRes{Message: "STATE_UPDATED"}, nil
}

func (s *State) updateBalances(ctx context.Context, dbTx *sqlx.Tx, txs []*proto.Transaction, pb *proto.Block) error {
	for _, tx := range txs {
		fromAcc, err := s.accModel.UpsertNoUpdate(ctx, dbTx, tx.GetFromAddr())
		if err != nil {
			return fmt.Errorf("failed to upsert src account %w", err)
		}

		toAcc, err := s.accModel.UpsertNoUpdate(ctx, dbTx, tx.GetToAddr())
		if err != nil {
			return fmt.Errorf("failed to upsert dest account %w", err)
		}

		fromAccBc := balancechange.BalanceChange{
			PreviousBalance: fromAcc.Balance,
			NewBalance:      fromAcc.Balance - tx.GetAmount() - tx.GetFee(),
			ChangeAmount:    -(tx.GetAmount() + tx.GetFee()),
			AccountID:       fromAcc.ID,
			BlockHeight:     pb.GetHeight(),
			BlockHash:       pb.GetHash(),
			TxHash:          tx.GetHash(),
			Timestamp:       tx.GetTimestamp(),
		}

		if err := s.balanceChangeModel.Crete(ctx, dbTx, fromAccBc); err != nil {
			return fmt.Errorf("failed to create source acc balance change record %w", err)
		}

		if _, err := s.accModel.Upsert(ctx, dbTx, account.Account{
			Address: fromAcc.Address,
			Balance: fromAccBc.NewBalance,
			Nonce:   fromAcc.Nonce + 1,
		}); err != nil {
			return fmt.Errorf("failed to update source account balance %w", err)
		}

		toAccBc := balancechange.BalanceChange{
			PreviousBalance: toAcc.Balance,
			NewBalance:      toAcc.Balance + tx.GetAmount(),
			ChangeAmount:    tx.GetAmount(),
			AccountID:       toAcc.ID,
			BlockHeight:     pb.GetHeight(),
			BlockHash:       pb.GetHash(),
			TxHash:          tx.GetHash(),
			Timestamp:       tx.GetTimestamp(),
		}

		if err := s.balanceChangeModel.Crete(ctx, dbTx, toAccBc); err != nil {
			return fmt.Errorf("failed to create destination acc balance change record %w", err)
		}

		if _, err := s.accModel.Upsert(ctx, dbTx, account.Account{
			Address: toAcc.Address,
			Balance: toAccBc.NewBalance,
			Nonce:   toAcc.Nonce + 1,
		}); err != nil {
			return fmt.Errorf("failed to update destination account balance %w", err)
		}
	}

	return nil
}

func (s *State) createBlock(ctx context.Context, dbTx *sqlx.Tx, txs []*proto.Transaction, pb *proto.Block) error {
	txsJson, err := json.Marshal(txs)
	if err != nil {
		return fmt.Errorf("createBlock failed to Marshal txs %w", err)
	}

	blockPld := block.BlockDB{
		Block: block.Block{
			Hash:       pb.GetHash(),
			PrevHash:   pb.GetPrevHash(),
			MerkleRoot: pb.GetMerkleRoot(),
			Timestamp:  pb.GetTimestamp(),
			Height:     pb.GetHeight(),
			Nonce:      pb.GetNonce(),
			// Difficulty: pb.GetDifficulty(),
		},
		TransactionsDB: string(txsJson),
	}

	if err := s.blockModel.SaveWithTX(ctx, dbTx, blockPld); err != nil {
		return fmt.Errorf("failed persisting block data %w, height: %v, block_hash: %v", err, blockPld.Height, blockPld.Hash)
	}

	receiptsPld := receipt.ProtoToReceipts(txs, pb.GetHash())
	if err := s.receiptModel.InsertBatch(ctx, dbTx, receiptsPld); err != nil {
		return fmt.Errorf("failed persisting receipts %w", err)
	}

	return nil
}

func (s *State) GetLatestBlock(ctx context.Context, in *proto.LastBlockReq) (*proto.LastBlockRes, error) {
	latestBlock, err := s.blockModel.GetLatest(ctx)
	if err != nil {
		s.log.Info("failed getting latest block data", "err", err)
		return nil, status.Error(codes.Internal, "failed getting latest block data")
	}

	return &proto.LastBlockRes{
		Block: block.ToProtoBlock(latestBlock),
	}, nil
}

func (s *State) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.apiPort))
	if err != nil {
		return fmt.Errorf("failed starting net listener %w", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	proto.RegisterStateServiceServer(server, s)

	s.log.Info("rpc server started", "port exposed", s.apiPort)
	return server.Serve(listener)
}
