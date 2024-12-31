package balancechange

import (
	"context"

	"com.perkunas/internal/db"
	"github.com/jmoiron/sqlx"
)

type Model struct {
	DB *db.DB
}

func (am *Model) Crete(ctx context.Context, db *sqlx.Tx, bc BalanceChange) error {
	query := `
		INSERT INTO balance_changes (account_id, previous_balance, new_balance, change_amount, block_height, block_hash, tx_hash)
		VALUES (:account_id, :previous_balance, :new_balance, :change_amount, :block_height, :block_hash, :tx_hash)
	`
	_, err := db.NamedExecContext(ctx, query, bc)
	return err
}
