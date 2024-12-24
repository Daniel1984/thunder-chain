package transaction

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidPublicKey = errors.New("invalid public key")
	ErrSigningError     = errors.New("signing error")
	ErrInvalidDataLen   = errors.New("invalid data length")
)

type Transaction struct {
	ID        string `json:"id" db:"id"`
	From      string `json:"from_addr" db:"from_addr"` // Sender's public key
	To        string `json:"to_addr" db:"to_addr"`     // Recipient's public key
	Signature string `json:"signature" db:"signature"` // Ed25519 signature
	Amount    uint64 `json:"amount" db:"amount"`
	Fee       uint64 `json:"fee" db:"fee"`
	Timestamp int64  `json:"timestamp" db:"timestamp"`
	Expires   int64  `json:"expires" db:"expires"`
}

func NewTransaction(from string, to string, amount, fee uint64, expiration time.Duration) *Transaction {
	return &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Timestamp: time.Now().Unix(),
		Expires:   time.Now().Add(expiration).Unix(),
	}
}

func (t *Transaction) CalculateHash() ([]byte, error) {
	hasher := sha256.New()

	hasher.Write([]byte(t.From))
	hasher.Write([]byte(t.To))

	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBytes, t.Amount)
	hasher.Write(amountBytes)

	feeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(feeBytes, t.Fee)
	hasher.Write(feeBytes)

	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(t.Timestamp))
	hasher.Write(timestampBytes)

	expiresBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(expiresBytes, uint64(t.Timestamp))
	hasher.Write(expiresBytes)

	return hasher.Sum(nil), nil
}

func (t *Transaction) Sign(privateKey ed25519.PrivateKey) error {
	publicKey := privateKey.Public().(ed25519.PublicKey)
	if string(publicKey) != t.From {
		return ErrInvalidPublicKey
	}

	messageHash, err := t.CalculateHash()
	if err != nil {
		return err
	}

	signature := ed25519.Sign(privateKey, messageHash)
	t.Signature = string(signature)
	return nil
}

func (t *Transaction) Verify() error {
	publicKey := ed25519.PublicKey([]byte(t.From))
	messageHash, err := t.CalculateHash()
	if err != nil {
		return err
	}

	if !ed25519.Verify(publicKey, messageHash, []byte(t.Signature)) {
		return ErrInvalidSignature
	}

	return nil
}
