package receipt

import (
	"encoding/json"
	"time"

	"com.perkunas/proto"
)

type Receipt struct {
	TxHash    string          `json:"txHash" db:"tx_hash"`
	BlockHash string          `json:"blockHash" db:"block_hash"`
	Status    string          `json:"status" db:"status"`
	GasUsed   int64           `json:"gasUsed" db:"gas_used"`
	Logs      json.RawMessage `json:"logs" db:"logs"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
}

func ProtoToReceipts(in []*proto.Transaction, blockHash string) []Receipt {
	var res []Receipt

	for _, tx := range in {
		res = append(res, Receipt{
			TxHash:    tx.GetHash(),
			BlockHash: blockHash,
			Status:    "ACCEPTED",
			Logs:      json.RawMessage(`[]`),
			// GasUsed: tx.GetGasUsed(),
		})
	}

	return res
}
