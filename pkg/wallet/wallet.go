package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"com.perkunas/internal/errmsg"
	"com.perkunas/internal/models/transaction"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

func New() (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	publicKey := &privateKey.PublicKey
	address := crypto.PubkeyToAddress(*publicKey).Hex()

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

func FromPrivateKey(privateKeyHex string) (*Wallet, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := &privateKey.PublicKey
	address := crypto.PubkeyToAddress(*publicKey).Hex()

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

func (w *Wallet) GetPublicKeyHex() string {
	return hex.EncodeToString(crypto.FromECDSAPub(w.PublicKey))
}

func (w *Wallet) GetPrivateKeyHex() string {
	return hex.EncodeToString(crypto.FromECDSA(w.PrivateKey))
}

func (w *Wallet) SignTransaction(tx *transaction.Transaction) error {
	if w.PrivateKey == nil {
		return errmsg.ErrSigningError
	}

	senderAddr := crypto.PubkeyToAddress(w.PrivateKey.PublicKey).Hex()
	if tx.From != senderAddr {
		return errmsg.ErrSignatureSenderMismatch
	}

	hash := crypto.Keccak256Hash(tx.CalculateHash())
	signature, err := crypto.Sign(hash.Bytes(), w.PrivateKey)
	if err != nil {
		return errmsg.ErrSigningError
	}

	tx.Signature = hex.EncodeToString(signature)
	return nil
}
