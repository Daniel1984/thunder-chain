package transaction

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	from := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	to := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	amount := uint64(100)
	fee := uint64(1)
	expiration := 10 * time.Minute

	tx, err := NewTransaction(from, to, amount, fee, expiration)
	assert.NoError(t, err)

	assert.Equal(t, from, tx.From)
	assert.Equal(t, to, tx.To)
	assert.Equal(t, amount, tx.Amount)
	assert.Equal(t, fee, tx.Fee)
	assert.LessOrEqual(t, time.Now().Unix(), tx.Timestamp)
	assert.Greater(t, tx.Expires, tx.Timestamp)
}

func TestCalculateHash(t *testing.T) {
	tx := &Transaction{
		From:      "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		To:        "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		Amount:    100,
		Fee:       1,
		Timestamp: time.Now().Unix(),
		Expires:   time.Now().Add(10 * time.Minute).Unix(),
	}

	hash, err := tx.CalculateHash()
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 32) // SHA-256 produces 32 bytes
}

func TestSignAndVerify(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	fromAddress := ConvertPublicKeyToAddress(publicKey)
	toAddress := "0x000000000000000000000000000000000000000a"

	tx, err := NewTransaction(fromAddress, toAddress, 100, 1, 10*time.Minute)
	assert.NoError(t, err)

	err = tx.Sign(privateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, tx.Signature)

	err = tx.Verify()
	assert.NoError(t, err)
}

func TestSignWithInvalidKey(t *testing.T) {
	publicKey1, _, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	_, privateKey2, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	fromAddress := ConvertPublicKeyToAddress(publicKey1)
	toAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"

	tx, err := NewTransaction(fromAddress, toAddress, 100, 1, 10*time.Minute)
	assert.NoError(t, err)

	err = tx.Sign(privateKey2)
	assert.ErrorIs(t, err, ErrInvalidPublicKey)
}

func TestVerifyWithTamperedData(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	pub2, _, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	fromAddress := ConvertPublicKeyToAddress(pub)
	toAddress := ConvertPublicKeyToAddress(pub2)

	tx, err := NewTransaction(fromAddress, toAddress, 100, 1, 10*time.Minute)
	assert.NoError(t, err)

	err = tx.Sign(priv)
	assert.NoError(t, err)

	tx.Amount = 200
	err = tx.Verify()
	assert.ErrorIs(t, err, ErrInvalidSignature)
}
