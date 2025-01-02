package account

import "com.perkunas/proto"

type Account struct {
	ID        int64  `json:"id" db:"id"`
	Address   string `json:"address" db:"address"`
	Balance   int64  `json:"balance" db:"balance"`
	Nonce     uint64 `json:"nonce" db:"nonce"`
	Timestamp int64  `json:"timestamp" db:"timestamp"`
}

func (acc Account) ToProto() *proto.Account {
	return &proto.Account{
		Id:        acc.ID,
		Address:   acc.Address,
		Balance:   acc.Balance,
		Nonce:     acc.Nonce,
		Timestamp: acc.Timestamp,
	}
}
