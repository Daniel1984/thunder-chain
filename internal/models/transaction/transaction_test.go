package transaction

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_CalculateHash(t *testing.T) {
	tx := &Transaction{
		From:      "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		To:        "0x7217d3eC0A0C357d7Dde4896094B83137c137E42",
		Amount:    1000,
		Fee:       10,
		Nonce:     1,
		Data:      "test data",
		Timestamp: time.Now().Unix(),
		Expires:   time.Now().Add(time.Hour).Unix(),
	}

	hash1 := tx.CalculateHash()
	assert.NotEmpty(t, hash1)

	hash2 := tx.CalculateHash()
	assert.Equal(t, hash1, hash2)

	// Different data should produce different hash
	tx.Amount = 2000
	hash3 := tx.CalculateHash()
	assert.NotEqual(t, hash1, hash3)
}

func TestTransaction_SetHash(t *testing.T) {
	tx := &Transaction{
		From:   "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		To:     "0x7217d3eC0A0C357d7Dde4896094B83137c137E42",
		Amount: 1000,
		Fee:    10,
		Nonce:  1,
	}

	tx.SetHash()
	assert.NotEmpty(t, tx.Hash)
	expected := hex.EncodeToString(tx.CalculateHash())
	assert.Equal(t, expected, tx.Hash)
}
