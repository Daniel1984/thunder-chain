package transaction

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_CalculateHash(t *testing.T) {
	tx := &Transaction{
		From:      "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		To:        "0x7217d3eC0A0C357d7Dde4896094B83137c137E42",
		Amount:    1000,
		Fee:       10,
		Nonce:     1,
		Data:      []byte("test data"),
		Timestamp: time.Now().Unix(),
		Expires:   time.Now().Add(time.Hour).Unix(),
	}

	hash1 := tx.CalculateHash()
	assert.NotEmpty(t, hash1)

	// Same data should produce same hash
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

func TestTransaction_SignAndVerify(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	tx := &Transaction{
		From:   address,
		To:     "0x7217d3eC0A0C357d7Dde4896094B83137c137E42",
		Amount: 1000,
		Fee:    10,
		Nonce:  1,
	}

	tx.SetHash()

	// Sign transaction
	err = SignTransaction(tx, privateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, tx.Signature)

	// Verify valid transaction
	err = tx.Verify()
	assert.NoError(t, err)

	// Test invalid hash
	originalHash := tx.Hash
	tx.Hash = "invalid"
	err = tx.Verify()
	assert.Equal(t, ErrInvalidHash, err)
	tx.Hash = originalHash

	// Test invalid signature
	tx.Signature[0] ^= 0x01
	err = tx.Verify()
	assert.Error(t, err)

	// Test invalid sender
	tx.From = "0x7217d3eC0A0C357d7Dde4896094B83137c137E42"
	err = tx.Verify()
	assert.Equal(t, ErrInvalidSignature, err)
}

func TestSignTransaction_InvalidKey(t *testing.T) {
	tx := &Transaction{
		From:   "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		To:     "0x7217d3eC0A0C357d7Dde4896094B83137c137E42",
		Amount: 1000,
	}

	err := SignTransaction(tx, nil)
	assert.Error(t, err)
}
