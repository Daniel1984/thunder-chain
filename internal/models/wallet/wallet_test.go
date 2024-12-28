package wallet

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	wallet, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	assert.Len(t, wallet.PublicKey, ed25519.PublicKeySize)
	assert.Len(t, wallet.PrivateKey, ed25519.PrivateKeySize)

	// Test address format
	assert.True(t, len(wallet.Address) == 42)
	assert.True(t, wallet.Address[:2] == "0x")

	// Test hex conversions
	pubHex := wallet.GetPublicKeyHex()
	assert.Len(t, pubHex, ed25519.PublicKeySize*2)

	privHex := wallet.GetPrivateKeyHex()
	assert.Len(t, privHex, ed25519.PrivateKeySize*2)
}
