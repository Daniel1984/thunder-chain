package transaction

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"com.perkunas/internal/errmsg"
	"com.perkunas/proto"
	"github.com/ethereum/go-ethereum/crypto"
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
		return errmsg.ErrInvalidSignatureFormat
	}

	msgHash := crypto.Keccak256Hash(t.CalculateHash())
	pubKey, err := crypto.Ecrecover(msgHash.Bytes(), sigBytes)
	if err != nil {
		return errmsg.ErrSignatureRecoveryFailed
	}

	publicKeyECDSA, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return errmsg.ErrInvalidPublicKeyFormat
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	if address != t.From {
		return errmsg.ErrSignatureSenderMismatch
	}

	// Then verify hash matches data
	calculatedHash := hex.EncodeToString(t.CalculateHash())
	if calculatedHash != t.Hash {
		return errmsg.ErrInvalidHash
	}

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

func ToProtoTx(in Transaction) *proto.Transaction {
	return &proto.Transaction{
		Hash:      in.Hash,
		FromAddr:  in.From,
		ToAddr:    in.To,
		Signature: in.Signature,
		Amount:    in.Amount,
		Fee:       in.Fee,
		Nonce:     in.Nonce,
		Data:      in.Data,
		Timestamp: in.Timestamp,
		Expires:   in.Expires,
	}
}
