package genesisblock

import (
	"context"

	"com.perkunas/internal/db"
	"github.com/jmoiron/sqlx"
)

type Model struct {
	DB *db.DB
}

func (bm *Model) SaveWithTX(ctx context.Context, db *sqlx.Tx, b GenesisBlock) error {
	query := `
		INSERT INTO blocks (hash, prev_hash, merkle_root, timestamp, height, nonce, transactions)
		VALUES (:hash, :prev_hash, :merkle_root, :timestamp, :height, :nonce, :transactions)
	`
	_, err := db.NamedExecContext(ctx, query, b.BlockDB)
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
