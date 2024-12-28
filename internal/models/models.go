package models

import (
	"com.perkunas/internal/models/transaction"
	"com.perkunas/internal/sqlite"
)

type Models struct {
	TransactionModel transaction.TransactionModel
}

func NewModels(db sqlite.Database) *Models {
	return &Models{
		TransactionModel: transaction.TransactionModel{DB: db},
	}
}
