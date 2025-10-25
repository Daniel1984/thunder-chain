package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	wallet, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	// Test that keys are not nil
	assert.NotNil(t, wallet.PublicKey)
	assert.NotNil(t, wallet.PrivateKey)

	// Test address format
	assert.True(t, len(wallet.Address) == 42)
	assert.True(t, wallet.Address[:2] == "0x")

	// Test hex conversions
	pubHex := wallet.GetPublicKeyHex()
	// ECDSA public key is 65 bytes (130 hex chars) when uncompressed with 0x04 prefix
	assert.Len(t, pubHex, 130)

	privHex := wallet.GetPrivateKeyHex()
	// ECDSA private key is 32 bytes (64 hex chars)
	assert.Len(t, privHex, 64)
}
