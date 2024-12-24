package transaction

import (
	"context"

	"com.perkunas/internal/sqlite"
)

type TransactionModel struct {
	DB sqlite.Database
}

func (tm *TransactionModel) Save(ctx context.Context, tx Transaction) error {
	query := `
		INSERT INTO mempool (id, from_addr, to_addr, signature, fee, amount, timestamp, expires)
		VALUES (:id, :from_addr, :to_addr, :signature, :fee, :amount, :timestamp, :expires)
	`
	_, err := tm.DB.NamedExecContext(ctx, query, tx)
	return err
}

func (tm *TransactionModel) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM mempool
		WHERE id=:id
	`
	_, err := tm.DB.NamedExecContext(ctx, query, map[string]interface{}{"id": id})
	return err
}
