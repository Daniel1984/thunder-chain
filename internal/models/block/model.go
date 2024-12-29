package block

import (
	"context"

	"com.perkunas/internal/sqlite"
)

type BlockModel struct {
	DB sqlite.Database
}

// TODO: persist transations in separate table
func (bm *BlockModel) Save(ctx context.Context, b BlockDB) error {
	query := `
		INSERT INTO blocks (hash, prev_hash, merkle_root, timestamp, height, nonce, transactions)
		VALUES (:hash, :prev_hash, :merkle_root, :timestamp, :height, :nonce, :transactions)
	`
	_, err := bm.DB.NamedExecContext(ctx, query, b)
	return err
}

func (bm *BlockModel) HasGenesisBlock(ctx context.Context) (bool, error) {
	query := `
		SELECT COUNT(*) FROM blocks
		WHERE height = ?
	`
	var count int
	err := bm.DB.GetContext(ctx, &count, query, 0)
	return count > 0, err
}
