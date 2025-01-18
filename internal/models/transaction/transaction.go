package transaction

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"com.perkunas/proto"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrInvalidSignature        = errors.New("invalid signature")
	ErrInvalidPublicKey        = errors.New("invalid public key")
	ErrSigningError            = errors.New("signing error")
	ErrInvalidDataLen          = errors.New("invalid data length")
	ErrInvalidHash             = errors.New("invalid transaction hash")
	ErrInvalidSignatureFormat  = errors.New("invalid signature format")
	ErrSignatureRecoveryFailed = errors.New("failed to recover public key from signature")
	ErrInvalidPublicKeyFormat  = errors.New("invalid public key format")
	ErrSignatureSenderMismatch = errors.New("signature does not match sender address")
)

type Transaction struct {
	ID        int64  `json:"id" db:"id"`
	Hash      string `json:"hash" db:"hash"`
	From      string `json:"from_addr" db:"from_addr"` // Sender's public key
	To        string `json:"to_addr" db:"to_addr"`     // Recipient's public key
	Data      string `json:"data,omitempty"`
	Signature string `json:"signature"`
	Amount    int64  `json:"amount" db:"amount"`
	Fee       int64  `json:"fee" db:"fee"`
	Nonce     uint64 `json:"nonce" db:"nonce"`
	Timestamp int64  `json:"timestamp" db:"timestamp"`
	Expires   int64  `json:"expires" db:"expires"`
}

func (t *Transaction) CalculateHash() []byte {
	hasher := sha256.New()
	buf := make([]byte, 8)

	// Order matters for consistency across nodes
	hasher.Write([]byte(t.From))
	hasher.Write([]byte(t.To))

	binary.Write(hasher, binary.BigEndian, t.Amount)
	binary.Write(hasher, binary.BigEndian, t.Fee)

	binary.BigEndian.PutUint64(buf, t.Nonce)
	hasher.Write(buf)

	return hasher.Sum(nil)
}

func (t *Transaction) SetHash() {
	t.Hash = hex.EncodeToString(t.CalculateHash())
}

func (t *Transaction) Verify() error {
	// Verify signature and sender first
	sigBytes, err := hex.DecodeString(t.Signature)
	if err != nil {
		return ErrInvalidSignatureFormat
	}

	msgHash := crypto.Keccak256Hash(t.CalculateHash())
	pubKey, err := crypto.Ecrecover(msgHash.Bytes(), sigBytes)
	if err != nil {
		return ErrSignatureRecoveryFailed
	}

	publicKeyECDSA, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return ErrInvalidPublicKeyFormat
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	if address != t.From {
		return ErrSignatureSenderMismatch
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

	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	if tx.From != address {
		return ErrSignatureSenderMismatch
	}

	hash := crypto.Keccak256Hash(tx.CalculateHash())
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return ErrSigningError
	}

	tx.Signature = hex.EncodeToString(signature)
	return nil
}

func ToProtoTxs(in []*Transaction) (out []*proto.Transaction) {
	for _, tx := range in {
		out = append(out, &proto.Transaction{
			Id:        tx.ID,
			Hash:      tx.Hash,
			FromAddr:  tx.From,
			ToAddr:    tx.To,
			Signature: tx.Signature,
			Amount:    tx.Amount,
			Fee:       tx.Fee,
			Nonce:     tx.Nonce,
			Data:      tx.Data,
			Timestamp: tx.Timestamp,
			Expires:   tx.Expires,
		})
	}

	return out
}

func FromProtoTxs(in []*proto.Transaction) (out []*Transaction) {
	for _, tx := range in {
		out = append(out, &Transaction{
			ID:        tx.Id,
			Hash:      tx.Hash,
			From:      tx.FromAddr,
			To:        tx.ToAddr,
			Signature: tx.Signature,
			Amount:    tx.Amount,
			Fee:       tx.Fee,
			Nonce:     tx.Nonce,
			Data:      tx.Data,
			Timestamp: tx.Timestamp,
			Expires:   tx.Expires,
		})
	}

	return out
}

func FromProtoTx(in *proto.Transaction) Transaction {
	return Transaction{
		Hash:      in.GetHash(),
		From:      in.GetFromAddr(),
		To:        in.GetToAddr(),
		Signature: in.GetSignature(),
		Amount:    in.GetAmount(),
		Fee:       in.GetFee(),
		Nonce:     in.GetNonce(),
		Data:      in.GetData(),
		Timestamp: in.GetTimestamp(),
		Expires:   in.GetExpires(),
	}
}
