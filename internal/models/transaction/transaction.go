package transaction

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
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

func validateAddress(addr string) bool {
	if !strings.HasPrefix(addr, "0x") {
		return false
	}
	if len(addr) != 42 { // 0x + 40 hex chars
		return false
	}
	_, err := hex.DecodeString(addr[2:])
	return err == nil
}

func NewTransaction(from, to string, amount, fee uint64, expiration time.Duration) (*Transaction, error) {
	if !validateAddress(from) || !validateAddress(to) {
		return nil, errors.New("invalid address format")
	}

	return &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Timestamp: time.Now().Unix(),
		Expires:   time.Now().Add(expiration).Unix(),
	}, nil
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

// ConvertPublicKeyToAddress converts an Ed25519 public key to a readable address.
func ConvertPublicKeyToAddress(publicKey ed25519.PublicKey) string {
	hash := sha256.Sum256(publicKey)            // Hash the public key
	address := hex.EncodeToString(hash[:])[:40] // Take the first 20 bytes (40 hex characters)
	return "0x" + strings.ToLower(address)
}

// Store public key hex during signing
func (t *Transaction) Sign(privateKey ed25519.PrivateKey) error {
	publicKey := privateKey.Public().(ed25519.PublicKey)
	if ConvertPublicKeyToAddress(publicKey) != t.From {
		return ErrInvalidPublicKey
	}

	t.ID = hex.EncodeToString(publicKey) // Store public key for verification
	messageHash, err := t.CalculateHash()
	if err != nil {
		return err
	}

	signature := ed25519.Sign(privateKey, messageHash)
	t.Signature = hex.EncodeToString(signature)
	return nil
}

// Use stored public key hex for verification
func (t *Transaction) Verify() error {
	messageHash, err := t.CalculateHash()
	if err != nil {
		return err
	}

	signature, err := hex.DecodeString(t.Signature)
	if err != nil {
		return err
	}

	pubKey, err := hex.DecodeString(t.ID) // Use stored public key
	if err != nil {
		return err
	}

	if !ed25519.Verify(ed25519.PublicKey(pubKey), messageHash, signature) {
		return ErrInvalidSignature
	}
	return nil
}
