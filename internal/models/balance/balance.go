package balance

type BalanceChange struct {
	ID              int64  `db:"id"`
	AccountID       int64  `db:"account_id"`
	PreviousBalance uint64 `db:"previous_balance"`
	NewBalance      uint64 `db:"new_balance"`
	ChangeAmount    int64  `db:"change_amount"` // Can be negative for outgoing transactions
	BlockHeight     uint64 `db:"block_height"`
	BlockHash       string `db:"block_hash"`
	TxHash          string `db:"tx_hash"`
	Timestamp       int64  `db:"timestamp"`
}
