package receipt

import (
	"context"

	"com.perkunas/internal/db"
	"github.com/jmoiron/sqlx"
)

type Model struct {
	DB *db.DB
}

func (am *Model) InsertBatch(ctx context.Context, db *sqlx.Tx, in []Receipt) error {
	query := `
		INSERT INTO balance_changes (tx_hash, block_hash, status, gas_used, logs)
		VALUES (:tx_hash, :block_hash, :status, :gas_used, :logs)
	`

	_, err := db.NamedExecContext(ctx, query, in)
	return err
}
