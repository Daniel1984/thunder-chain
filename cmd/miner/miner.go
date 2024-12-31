package main

import (
	"com.perkunas/internal/models/block"
	"com.perkunas/internal/models/transaction"
)

type Miner struct {
	// mempool    MempoolClient // gRPC client to fetch pending transactions
	// stateDB    StateDB       // To verify transaction validity
	// blockDB    BlockDB       // To access/store blocks
	difficulty uint64 // Current mining difficulty
	reward     uint64 // Block reward amount
}

type MiningCandidate struct {
	PrevBlock  *block.Block
	Txs        []*transaction.Transaction
	Difficulty uint64
	Nonce      uint64
	Timestamp  int64
}

// func (m *Miner) Start(ctx context.Context) error {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil

// 		default:
// 			// 1. Get pending transactions from mempool
// 			txs, err := m.mempool.GetPendingTransactions(ctx)
// 			if err != nil {
// 				continue
// 			}

// 			// 2. Get latest block
// 			prevBlock, err := m.blockDB.GetLatestBlock(ctx)
// 			if err != nil {
// 				continue
// 			}

// 			// 3. Create candidate block
// 			candidate := &MiningCandidate{
// 				PrevBlock:  prevBlock,
// 				Timestamp:  time.Now().Unix(),
// 				Txs:        txs,
// 				Difficulty: m.getCurrentDifficulty(),
// 			}

// 			// 4. Mine block (find valid nonce)
// 			newBlock, err := m.mineBlock(ctx, candidate)
// 			if err != nil {
// 				continue
// 			}

// 			// 5. Submit mined block
// 			if err := m.submitBlock(ctx, newBlock); err != nil {
// 				continue
// 			}
// 		}
// 	}
// }

// func (m *Miner) mineBlock(ctx context.Context, candidate *MiningCandidate) (*block.Block, error) {
// 	// 1. Validate transactions
// 	validTxs, err := m.validateTransactions(ctx, candidate.Txs)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. Create block with mining reward
// 	block := &block.Block{
// 		PrevHash:     candidate.PrevBlock.Hash,
// 		Height:       candidate.PrevBlock.Height + 1,
// 		Timestamp:    candidate.Timestamp,
// 		Difficulty:   candidate.Difficulty,
// 		Transactions: append(validTxs, createRewardTx(m.reward)),
// 	}

// 	// 3. Find valid nonce (Proof of Work)
// 	for nonce := uint64(0); ; nonce++ {
// 		select {
// 		case <-ctx.Done():
// 			return nil, ctx.Err()
// 		default:
// 			block.Nonce = nonce
// 			block.CalculateHash()

// 			if isHashValid(block.Hash, block.Difficulty) {
// 				return block, nil
// 			}
// 		}
// 	}
// }

// func (m *Miner) submitBlock(ctx context.Context, block *block.Block) error {
// 	// 1. Save block to database
// 	if err := m.blockDB.SaveBlock(ctx, block); err != nil {
// 		return err
// 	}

// 	// 2. Update account balances
// 	if err := m.updateState(ctx, block); err != nil {
// 		return err
// 	}

// 	// 3. Remove mined transactions from mempool
// 	if err := m.mempool.RemoveTransactions(ctx, block.Transactions); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (m *Miner) updateState(ctx context.Context, block *block.Block) error {
// 	// Begin transaction
// 	tx, err := m.stateDB.BeginTx(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	// Apply all transactions
// 	for _, transaction := range block.Transactions {
// 		// Deduct from sender
// 		if err := tx.UpdateBalance(transaction.From, -transaction.Amount); err != nil {
// 			return err
// 		}
// 		// Add to recipient
// 		if err := tx.UpdateBalance(transaction.To, transaction.Amount); err != nil {
// 			return err
// 		}
// 	}

// 	return tx.Commit()
// }

// func isHashValid(hash string, difficulty uint64) bool {
// 	// Convert difficulty to required leading zeros
// 	prefix := strings.Repeat("0", int(difficulty))

// 	// Check if hash starts with required number of zeros
// 	return strings.HasPrefix(hash, prefix)
// }
