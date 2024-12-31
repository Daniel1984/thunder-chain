package account

import (
	"context"

	"com.perkunas/internal/db"
)

type Model struct {
	DB db.DB
}

func (am *Model) Upsert(ctx context.Context, a Account) (Account, error) {
	query := `
		INSERT INTO blocks (address, balance, nonce, timestamp)
		VALUES (:address, :balance, :nonce, :timestamp)
		ON CONFLICT (address) DO UPDATE SET balance = excluded.balance, nonce = excluded.nonce
		RETURNING id, address, balance, nonce
	`

	var acc Account
	if err := am.DB.WriteDB.QueryRowContext(ctx, query, a).Scan(
		&acc.ID,
		&acc.Address,
		&acc.Balance,
		&acc.Nonce,
	); err != nil {
		return Account{}, err
	}

	return acc, nil
}
