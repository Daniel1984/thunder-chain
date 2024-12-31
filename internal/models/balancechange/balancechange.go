package balancechange

type BalanceChange struct {
	ID              int64  `json:"id" db:"-"`
	AccountID       int64  `json:"account_id" db:"account_id"`
	PreviousBalance int64  `json:"previous_balance" db:"previous_balance"`
	NewBalance      int64  `json:"new_balance" db:"new_balance"`
	ChangeAmount    int64  `json:"change_amount" db:"change_amount"`
	Timestamp       int64  `json:"timestamp" db:"timestamp"`
	BlockHeight     uint64 `json:"block_height" db:"block_height"`
	BlockHash       string `json:"block_hash" db:"block_hash"`
	TxHash          string `json:"tx_hash" db:"tx_hash"`
}
