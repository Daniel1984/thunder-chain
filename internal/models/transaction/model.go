package transaction

import (
	"context"

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

func (tm *Model) Delete(ctx context.Context, hash string) error {
	query := `
		DELETE FROM mempool
		WHERE hash=:hash
	`
	_, err := tm.DB.WriteDB.NamedExecContext(ctx, query, map[string]interface{}{"hash": hash})
	return err
}

func (tm *Model) Pending(ctx context.Context) ([]Transaction, error) {
	query := `
		SELECT
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
