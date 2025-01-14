package transaction

import (
	"context"
	"fmt"
	"strings"

	"com.perkunas/internal/db"
)

type Model struct {
	DB *db.DB
}

func (tm *Model) Save(ctx context.Context, tx Transaction) error {
	query := `
		INSERT INTO mempool (hash, from_addr, to_addr, signature, fee, amount, nonce, timestamp, expires)
		VALUES (:hash, :from_addr, :to_addr, :signature, :fee, :amount, :nonce, :timestamp, :expires)
	`
	_, err := tm.DB.WriteDB.NamedExecContext(ctx, query, tx)
	return err
}

func (tm *Model) DeleteBatch(ctx context.Context, IDs []int64) error {
	if len(IDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(IDs))
	args := make([]interface{}, len(IDs))
	for i, id := range IDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		DELETE FROM mempool
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	_, err := tm.DB.WriteDB.ExecContext(ctx, query, args...)
	return err
}

func (tm *Model) Pending(ctx context.Context) ([]Transaction, error) {
	query := `
		SELECT
			id,
			hash,
			from_addr,
			to_addr,
			signature,
			fee,
			amount,
			nonce,
			timestamp,
			expires
		FROM mempool
		ORDER BY fee DESC LIMIT 2000
	`

	var res []Transaction
	if err := tm.DB.ReadDB.Select(&res, query); err != nil {
		return nil, err
	}

	return res, nil
}
