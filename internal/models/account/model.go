package account

import (
	"context"

	"com.perkunas/internal/db"
	"github.com/jmoiron/sqlx"
)

type Model struct {
	DB *db.DB
}

func (am *Model) Upsert(ctx context.Context, db *sqlx.Tx, a Account) (Account, error) {
	query := `
		INSERT INTO accounts (address, balance, nonce)
		VALUES (?, ?, ?)
		ON CONFLICT (address) DO UPDATE SET balance = excluded.balance, nonce = excluded.nonce
		RETURNING id, address, balance, nonce
	`

	var acc Account
	if err := db.QueryRowContext(ctx, query, a.Address, a.Balance, a.Nonce).Scan(
		&acc.ID,
		&acc.Address,
		&acc.Balance,
		&acc.Nonce,
	); err != nil {
		return Account{}, err
	}

	return acc, nil
}

func (am *Model) UpsertNoUpdate(ctx context.Context, db *sqlx.Tx, addr string) (Account, error) {
	query := `
		INSERT INTO accounts (address)
		VALUES (?)
		ON CONFLICT (address) DO UPDATE SET address = excluded.address
		RETURNING id, address, balance, nonce
	`

	var acc Account
	if err := db.QueryRowContext(ctx, query, addr).Scan(
		&acc.ID,
		&acc.Address,
		&acc.Balance,
		&acc.Nonce,
	); err != nil {
		return Account{}, err
	}

	return acc, nil
}

func (am *Model) Get(ctx context.Context, addr string) (Account, error) {
	query := `
		SELECT id, address, balance, nonce, timestamp
		FROM accounts
		WHERE address = ?
	`

	var acc Account
	if err := am.DB.ReadDB.Get(&acc, query, addr); err != nil {
		return acc, err
	}

	return acc, nil
}
