package wallet

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
)

type Wallet struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Address    string
}

func New() (*Wallet, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(pub)
	address := "0x" + hex.EncodeToString(hash[:20])

	return &Wallet{
		PrivateKey: priv,
		PublicKey:  pub,
		Address:    address,
	}, nil
}

func (w *Wallet) GetPublicKeyHex() string {
	return hex.EncodeToString(w.PublicKey)
}

func (w *Wallet) GetPrivateKeyHex() string {
	return hex.EncodeToString(w.PrivateKey)
}
