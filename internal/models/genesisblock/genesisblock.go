package genesisblock

import (
	"com.perkunas/internal/models/account"
	"com.perkunas/internal/models/block"
)

type GenesisBlock struct {
	block.BlockDB
	Accounts []account.Account `json:"accounts"`
}
