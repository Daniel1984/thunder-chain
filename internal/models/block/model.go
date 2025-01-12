package block

import (
	"context"

	"com.perkunas/internal/db"
	"github.com/jmoiron/sqlx"
)

type Model struct {
	DB *db.DB
}

func (bm *Model) Save(ctx context.Context, b BlockDB) error {
	query := `
		INSERT INTO blocks (hash, prev_hash, merkle_root, timestamp, height, nonce, transactions)
		VALUES (:hash, :prev_hash, :merkle_root, :timestamp, :height, :nonce, :transactions)
	`
	_, err := bm.DB.WriteDB.NamedExecContext(ctx, query, b)
	return err
}

func (bm *Model) SaveWithTX(ctx context.Context, db *sqlx.Tx, b BlockDB) error {
	query := `
		INSERT INTO blocks (hash, prev_hash, merkle_root, timestamp, height, nonce, transactions)
		VALUES (:hash, :prev_hash, :merkle_root, :timestamp, :height, :nonce, :transactions)
	`
	_, err := db.NamedExecContext(ctx, query, b)
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

func (bm *Model) GetLatest(ctx context.Context) (Block, error) {
	query := `
		SELECT
			hash,
			prev_hash,
			merkle_root,
			height,
			nonce,
			difficulty,
			timestamp
		FROM blocks
		ORDER BY height DESC LIMIT 1"
	`

	var res Block
	return res, bm.DB.ReadDB.Get(&res, query)
}
