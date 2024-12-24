package block

import (
	"crypto/sha256"
	"encoding/binary"
	"time"

	"com.perkunas/internal/models/transaction"
)

type Block struct {
	Hash         string
	PrevHash     string
	MerkleRoot   string
	Timestamp    int64
	Height       uint64
	Nonce        uint64
	Transactions []*transaction.Transaction
}

func NewBlock() *Block {
	return &Block{
		Timestamp:    time.Now().Unix(),
		Transactions: make([]*transaction.Transaction, 0),
	}
}

func (b *Block) CalculateHash() (string, error) {
	merkleRoot, err := b.CalculateMerkleRoot()
	if err != nil {
		return "", err
	}
	b.MerkleRoot = merkleRoot

	hasher := sha256.New()
	hasher.Write([]byte(b.PrevHash))
	binary.Write(hasher, binary.LittleEndian, b.Timestamp)
	binary.Write(hasher, binary.LittleEndian, b.Height)
	binary.Write(hasher, binary.LittleEndian, b.Nonce)
	hasher.Write([]byte(b.MerkleRoot))

	return string(hasher.Sum(nil)), nil
}

func hashPair(left, right string) (string, error) {
	hasher := sha256.New()
	hasher.Write([]byte(left))
	hasher.Write([]byte(right))
	return string(hasher.Sum(nil)), nil
}

func (b *Block) AddTransaction(transaction *transaction.Transaction) error {
	b.Transactions = append(b.Transactions, transaction)
	merkleRoot, err := b.CalculateMerkleRoot()
	if err != nil {
		return err
	}

	b.MerkleRoot = merkleRoot
	return nil
}

func (b *Block) CalculateMerkleRoot() (string, error) {
	if len(b.Transactions) == 0 {
		return "", nil
	}

	currentLevel := make([]string, 0)

	for _, tx := range b.Transactions {
		hash, err := tx.CalculateHash()
		if err != nil {
			return "", err
		}
		currentLevel = append(currentLevel, string(hash))
	}

	// If odd number of transactions, duplicate last one
	if len(currentLevel)%2 == 1 {
		currentLevel = append(currentLevel, currentLevel[len(currentLevel)-1])
	}

	// Keep hashing pairs until we get to the root
	for len(currentLevel) > 1 {
		nextLevel := make([]string, 0)

		for i := 0; i < len(currentLevel); i += 2 {
			combined, err := hashPair(currentLevel[i], currentLevel[i+1])
			if err != nil {
				return "", err
			}
			nextLevel = append(nextLevel, combined)
		}

		currentLevel = nextLevel

		if len(currentLevel)%2 == 1 && len(currentLevel) > 1 {
			currentLevel = append(currentLevel, currentLevel[len(currentLevel)-1])
		}
	}

	return currentLevel[0], nil
}
