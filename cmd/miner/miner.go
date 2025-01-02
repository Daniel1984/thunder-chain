package main

import (
	"context"
	"errors"
	"strings"

	"com.perkunas/internal/models/block"
	"com.perkunas/internal/models/transaction"
	"com.perkunas/proto"
)

type MiningCandidate struct {
	PrevBlock  *proto.Block
	Txs        []*transaction.Transaction
	Difficulty uint64
	Nonce      uint64
	Timestamp  int64
}

func (app *App) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			// 1. get pending transactions from mempool
			txsRes, err := app.mempoolRPC.PendingTransactions(ctx, nil)
			if err != nil {
				app.log.Error("failed getting pending transactions", "err", err)
				continue
			}

			txs := txsRes.GetTransactions()
			if len(txs) == 0 {
				app.log.Info("notransactions in mempool")
				continue
			}

			// 2. get latest block
			prevBlock, err := app.blocksRPC.GetLatestBlock(ctx, nil)
			if err != nil {
				app.log.Error("failed gettin latest block", "err", err)
				continue
			}

			// 3. create candidate block
			candidate := &MiningCandidate{
				PrevBlock: prevBlock.GetBlock(),
				Txs:       transaction.FromProtoTxs(txs),
				// Difficulty: app.getCurrentDifficulty(),
				// Timestamp:  time.Now().Unix(),
			}

			// 4. mine block (find valid nonce)
			newBlock, err := app.mineBlock(ctx, candidate)
			if err != nil {
				app.log.Error("failed to mine block", "err", err)
				continue
			}

			// 5. submit mined block
			if err := app.submitBlock(ctx, newBlock); err != nil {
				app.log.Error("failed to submit mined block", "err", err)
			}
		}
	}
}

func (app *App) mineBlock(ctx context.Context, candidate *MiningCandidate) (*block.Block, error) {
	// 1. validate transactions
	validTxs := app.validateTransactions(ctx, candidate.Txs)
	if len(validTxs) == 0 {
		return nil, errors.New("no valid transactions found")
	}

	// 2. create block TODO: with mining reward
	block := &block.Block{
		PrevHash:     candidate.PrevBlock.Hash,
		Height:       candidate.PrevBlock.Height + 1,
		Timestamp:    candidate.Timestamp,
		Transactions: validTxs,
		// Difficulty:   candidate.Difficulty,
		// Transactions: append(validTxs, createRewardTx(app.reward)),
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

func (app *App) validateTransactions(ctx context.Context, txs []*transaction.Transaction) []*transaction.Transaction {
	validTxs := make([]*transaction.Transaction, 0)

	// track used nonces to prevent double-spending within same block
	usedNonces := make(map[string]uint64) // address -> nonce

	for _, tx := range txs {
		// skip invalid transactions but continue processing others
		if err := tx.Verify(); err != nil {
			app.log.Warn("invalid transaction skipped", "hash", tx.Hash, "error", err)
			continue
		}

		// check sender balance
		fromAccountRes, err := app.stateRPC.GetAccountByAddress(ctx, &proto.GetAccountByAddressRequest{Address: tx.From})
		if err != nil {
			app.log.Error("failed getting account by address", "address", tx.From, "err", err)
			continue
		}

		fromAcc := fromAccountRes.GetAccount()
		if fromAcc == nil {
			app.log.Info("account not found by address", "addr", tx.From)
			continue
		}

		if fromAcc.GetBalance() < tx.Amount+tx.Fee {
			app.log.Info("insufficient balance", "addr", tx.From, "balance", fromAcc.GetBalance(), "amount", tx.Amount+tx.Fee)
			continue
		}

		// check nonce
		if tx.Nonce != fromAcc.GetNonce()+1 {
			app.log.Info("invalid tx nonce", "txNonce", tx.Nonce, "accNonce", fromAcc.GetNonce()+1)
			continue
		}

		if lastNonce, exists := usedNonces[tx.From]; exists && tx.Nonce <= lastNonce {
			app.log.Info("nonce is smaller or equal to previously used one", "txNonce", tx.Nonce, "lastUsedNonce", lastNonce)
			continue
		}

		usedNonces[tx.From] = tx.Nonce
		validTxs = append(validTxs, tx)
	}

	return validTxs
}

func (app *App) submitBlock(ctx context.Context, block *block.Block) error {
	// 1. Save block to database
	blockPld := &proto.CreateBlockRequest{
		Block: &proto.Block{
			Hash:         block.Hash,
			PrevHash:     block.PrevHash,
			MerkleRoot:   block.MerkleRoot,
			Height:       block.Height,
			Nonce:        block.Nonce,
			Transactions: transaction.ToProtoTxs(block.Transactions),
		},
	}
	if _, err := app.blocksRPC.CreateBlock(ctx, blockPld); err != nil {
		app.log.Error("failed to crete block", "err", err)
		return err
	}

	// 2. Update account balances
	// post to statechange RPC ?

	// 3. Remove mined transactions from mempool
	// if err := app.mempoolRPC.RemoveTransactions(ctx, block.Transactions); err != nil {
	// 	app.log.Error("failed to clear mempool transactions", "err", err)
	// 	return err
	// }

	return nil
}

func isHashValid(hash string, difficulty uint64) bool {
	// Convert difficulty to required leading zeros
	prefix := strings.Repeat("0", int(difficulty))

	// Check if hash starts with required number of zeros
	return strings.HasPrefix(hash, prefix)
}
