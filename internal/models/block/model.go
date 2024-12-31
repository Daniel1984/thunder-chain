package block

import (
	"context"

	"com.perkunas/internal/db"
)

type Model struct {
	DB *db.DB
}

// TODO: persist transations in separate table
func (bm *Model) Save(ctx context.Context, b BlockDB) error {
	query := `
		INSERT INTO blocks (hash, prev_hash, merkle_root, timestamp, height, nonce, transactions)
		VALUES (:hash, :prev_hash, :merkle_root, :timestamp, :height, :nonce, :transactions)
	`
	_, err := bm.DB.WriteDB.NamedExecContext(ctx, query, b)
	return err
}

func (bm *Model) HasGenesisBlock(ctx context.Context) (bool, error) {
	query := `
		SELECT COUNT(*) FROM blocks
		WHERE height = ?
	`
	var count int
	err := bm.DB.ReadDB.GetContext(ctx, &count, query, 0)
	return count > 0, err
}
