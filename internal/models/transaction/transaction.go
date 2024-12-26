package transaction

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidPublicKey = errors.New("invalid public key")
	ErrSigningError     = errors.New("signing error")
	ErrInvalidDataLen   = errors.New("invalid data length")
	ErrInvalidHash      = errors.New("invalid transaction hash")
)

type Transaction struct {
	Hash      string `json:"hash" db:"id"`
	From      string `json:"from_addr" db:"from_addr"` // Sender's public key
	To        string `json:"to_addr" db:"to_addr"`     // Recipient's public key
	Amount    uint64 `json:"amount" db:"amount"`
	Fee       uint64 `json:"fee" db:"fee"`
	Nonce     uint64 `json:"nonce"`
	Data      []byte `json:"data,omitempty"`
	Signature []byte `json:"signature"`
	Timestamp int64  `json:"timestamp" db:"timestamp"`
	Expires   int64  `json:"expires" db:"expires"`
}

func (t *Transaction) CalculateHash() []byte {
	hasher := sha256.New()
	buf := make([]byte, 8)

	// Order matters for consistency across nodes
	hasher.Write([]byte(t.From))
	hasher.Write([]byte(t.To))

	binary.BigEndian.PutUint64(buf, t.Amount)
	hasher.Write(buf)

	binary.BigEndian.PutUint64(buf, t.Fee)
	hasher.Write(buf)

	binary.BigEndian.PutUint64(buf, t.Nonce)
	hasher.Write(buf)

	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(t.Timestamp))
	hasher.Write(timestampBytes)

	expiresBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(expiresBytes, uint64(t.Expires))
	hasher.Write(expiresBytes)

	if len(t.Data) > 0 {
		hasher.Write(t.Data)
	}

	return hasher.Sum(nil)
}

func (t *Transaction) SetHash() {
	t.Hash = hex.EncodeToString(t.CalculateHash())
}

func (t *Transaction) Verify() error {
	// Verify signature and sender first
	msgHash := crypto.Keccak256Hash(t.CalculateHash())
	pubKey, err := crypto.Ecrecover(msgHash.Bytes(), t.Signature)
	if err != nil {
		return ErrInvalidSignature
	}

	publicKeyECDSA, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return ErrInvalidSignature
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	if address != t.From {
		return ErrInvalidSignature
	}

	// Then verify hash matches data
	calculatedHash := hex.EncodeToString(t.CalculateHash())
	if calculatedHash != t.Hash {
		return ErrInvalidHash
	}

	return nil
}

// Reference for client-side signing
func SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) error {
	if privateKey == nil {
		return ErrSigningError
	}

	hash := crypto.Keccak256Hash(tx.CalculateHash())
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return ErrSigningError
	}
	tx.Signature = signature
	return nil
}
