package errmsg

import "errors"

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
