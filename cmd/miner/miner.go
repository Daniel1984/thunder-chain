package main

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"com.perkunas/internal/models/block"
	"com.perkunas/internal/models/transaction"
	"com.perkunas/proto"
)

type Miner struct {
	log        *slog.Logger
	mempoolAPI string
	stateAPI   string
	mempoolRPC proto.MempoolServiceClient
	stateRPC   proto.StateServiceClient
}

type MiningCandidate struct {
	PrevBlock  *proto.Block
	Txs        []*transaction.Transaction
	Difficulty uint64
	Nonce      uint64
	Timestamp  int64
}

func (m *Miner) Start(ctx context.Context) error {
	for {
		time.Sleep(20 * time.Second)
		select {
		case <-ctx.Done():
			return nil

		default:
			// 1. get pending transactions from mempool
			pendTxs, err := m.mempoolRPC.PendingTransactions(ctx, nil)
			if err != nil {
				m.log.Error("failed getting pending transactions", "err", err)
				continue
			}

			txs := pendTxs.GetTransactions()
			if len(txs) == 0 {
				m.log.Info("no transactions in mempool")
				continue
			}

			// 2. get latest block
			prevBlock, err := m.stateRPC.GetLatestBlock(ctx, nil)
			if err != nil {
				m.log.Error("failed gettin latest block", "err", err)
				continue
			}

			// 3. create candidate block
			candidate := &MiningCandidate{
				PrevBlock: prevBlock.GetBlock(),
				Txs:       transaction.FromProtoTxs(txs),
				// Difficulty: m.getCurrentDifficulty(),
				// Timestamp:  time.Now().Unix(),
			}

			// 4. mine block (find valid nonce)
			newBlock, err := m.mineBlock(ctx, candidate)
			if err != nil {
				m.log.Error("failed to mine block", "err", err)
				continue
			}

			// 5. persist new block
			if err := m.persistBlock(ctx, newBlock); err != nil {
				m.log.Error("failed to update chain state", "err", err)
				continue
			}

			// 6. delete processed txs from mempool
			if err := m.deleteTxs(ctx, newBlock); err != nil {
				m.log.Error("failed to delete processed transactions", "err", err)
			}
		}
	}
}

func (m *Miner) mineBlock(ctx context.Context, mc *MiningCandidate) (*block.Block, error) {
	// 1. validate transactions
	validTxs := m.validateTransactions(ctx, mc.Txs)
	if len(validTxs) == 0 {
		return nil, errors.New("no valid transactions found")
	}

	// 2. create block TODO: with mining reward
	block := &block.Block{
		PrevHash:     mc.PrevBlock.Hash,
		Height:       mc.PrevBlock.Height + 1,
		Timestamp:    mc.Timestamp,
		Transactions: validTxs,
		// Difficulty:   mc.Difficulty,
		// Transactions: append(validTxs, createRewardTx(m.reward)),
	}

	// 3. find valid nonce (Proof of Work)
	for nonce := uint64(0); ; nonce++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			block.Nonce = nonce
			hash, err := block.CalculateHash()
			if err != nil {
				return nil, err
			}
			block.Hash = hash

			if isHashValid(block.Hash, block.Difficulty) {
				return block, nil
			}
		}
	}
}

func (m *Miner) validateTransactions(ctx context.Context, txs []*transaction.Transaction) []*transaction.Transaction {
	validTxs := make([]*transaction.Transaction, 0)

	// track used nonces to prevent double-spending within same block
	usedNonces := make(map[string]uint64) // address -> nonce

	for _, tx := range txs {
		// skip invalid transactions but continue processing others
		if err := tx.Verify(); err != nil {
			m.log.Warn("invalid transaction skipped", "hash", tx.Hash, "error", err)
			continue
		}

		// check sender balance
		fromAccountRes, err := m.stateRPC.GetAccountByAddress(ctx, &proto.AccountByAddressReq{Address: tx.From})
		if err != nil {
			m.log.Error("failed getting account by address", "address", tx.From, "err", err)
			continue
		}

		fromAcc := fromAccountRes.GetAccount()
		if fromAcc == nil {
			m.log.Info("account not found by address", "addr", tx.From)
			continue
		}

		if fromAcc.GetBalance() < tx.Amount+tx.Fee {
			m.log.Info("insufficient balance", "addr", tx.From, "balance", fromAcc.GetBalance(), "amount", tx.Amount+tx.Fee)
			continue
		}

		// check nonce
		if tx.Nonce != fromAcc.GetNonce()+1 {
			m.log.Info("invalid tx nonce", "txNonce", tx.Nonce, "accNonce", fromAcc.GetNonce()+1)
			continue
		}

		if lastNonce, exists := usedNonces[tx.From]; exists && tx.Nonce <= lastNonce {
			m.log.Info("nonce is smaller or equal to previously used one", "txNonce", tx.Nonce, "lastUsedNonce", lastNonce)
			continue
		}

		usedNonces[tx.From] = tx.Nonce
		validTxs = append(validTxs, tx)
	}

	return validTxs
}

func (m *Miner) persistBlock(ctx context.Context, b *block.Block) error {
	_, err := m.stateRPC.CreateBlock(ctx, &proto.CreateBlockReq{
		Block: &proto.Block{
			Hash:         b.Hash,
			Height:       b.Height,
			PrevHash:     b.PrevHash,
			MerkleRoot:   b.MerkleRoot,
			Timestamp:    b.Timestamp,
			Transactions: transaction.ToProtoTxs(b.Transactions),
		},
	})

	return err
}

func (m *Miner) deleteTxs(ctx context.Context, b *block.Block) error {
	var idsToDelete []int64
	for _, tx := range b.Transactions {
		idsToDelete = append(idsToDelete, tx.ID)
	}

	_, err := m.mempoolRPC.DeleteMempoolBatch(ctx, &proto.DeleteMempoolBatchRequest{Ids: idsToDelete})
	return err
}

func isHashValid(hash string, difficulty uint64) bool {
	// Convert difficulty to required leading zeros
	prefix := strings.Repeat("0", int(difficulty))

	// Check if hash starts with required number of zeros
	return strings.HasPrefix(hash, prefix)
}
