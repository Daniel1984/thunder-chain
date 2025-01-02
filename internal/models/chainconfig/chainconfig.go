package chainconfig

type ChainConfig struct {
	InitialDifficulty uint64
	BlockTime         uint64
	DifficultyAdjust  uint64
	MaxTxPerBlock     uint64
	BlockReward       uint64
}
