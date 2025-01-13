package block

import (
	"testing"
	"time"

	"com.perkunas/internal/models/transaction"
	"github.com/stretchr/testify/assert"
)

func TestNewBlock(t *testing.T) {
	block := NewBlock()
	assert.NotNil(t, block)
	assert.Empty(t, block.Hash)
	assert.Empty(t, block.PrevHash)
	assert.NotZero(t, block.Timestamp)
	assert.Zero(t, block.Height)
	assert.Zero(t, block.Nonce)
	assert.Empty(t, block.MerkleRoot)
	assert.Empty(t, block.Transactions)
}

func TestCalculateHash(t *testing.T) {
	block := NewBlock()
	block.PrevHash = "previous_hash"
	block.Height = 1
	block.Nonce = 12345

	hash1, err := block.CalculateHash()
	assert.NoError(t, err)
	assert.NotEmpty(t, hash1)

	// Test hash consistency
	hash2, err := block.CalculateHash()
	assert.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Test hash changes with different data
	block.Nonce++
	hash3, err := block.CalculateHash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}

func TestHashPair(t *testing.T) {
	left := "hash1"
	right := "hash2"

	hash1 := hashPair(left, right)
	assert.NotEmpty(t, hash1)

	// Test consistency
	hash2 := hashPair(left, right)
	assert.Equal(t, hash1, hash2)

	// Test different input produces different hash
	hash3 := hashPair(right, left)
	assert.NotEqual(t, hash1, hash3)
}

func TestAddTransaction(t *testing.T) {
	block := NewBlock()
	tx := &transaction.Transaction{
		From:      "sender",
		To:        "recipient",
		Amount:    100,
		Fee:       10,
		Timestamp: time.Now().Unix(),
		Expires:   time.Now().Add(time.Hour).Unix(),
	}

	err := block.AddTransaction(tx)
	assert.NoError(t, err)
	assert.Len(t, block.Transactions, 1)
	assert.NotEmpty(t, block.MerkleRoot)
}

func TestCalculateMerkleRoot(t *testing.T) {
	block := NewBlock()

	// Test empty block
	root, err := block.CalculateMerkleRoot()
	assert.Error(t, err)
	assert.Empty(t, root)

	// Test single transaction
	tx1 := &transaction.Transaction{
		From:      "sender1",
		To:        "recipient1",
		Amount:    100,
		Timestamp: time.Now().Unix(),
	}
	err = block.AddTransaction(tx1)
	assert.NoError(t, err)
	root1, err := block.CalculateMerkleRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, root1)

	// Test multiple transactions
	tx2 := &transaction.Transaction{
		From:      "sender2",
		To:        "recipient2",
		Amount:    200,
		Timestamp: time.Now().Unix(),
	}
	err = block.AddTransaction(tx2)
	assert.NoError(t, err)
	root2, err := block.CalculateMerkleRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, root2)
	assert.NotEqual(t, root1, root2)

	// Test merkle root changes with transaction modification
	block.Transactions[0].Amount = 300
	root3, err := block.CalculateMerkleRoot()
	assert.NoError(t, err)
	assert.NotEqual(t, root2, root3)
}
