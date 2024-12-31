package account

type Account struct {
	ID        int64  `json:"id" db:"id"`
	Address   string `json:"address" db:"address"`
	Balance   int64  `json:"balance" db:"balance"`
	Nonce     uint64 `json:"nonce" db:"nonce"`
	Timestamp int64  `json:"timestamp" db:"timestamp"`
}
