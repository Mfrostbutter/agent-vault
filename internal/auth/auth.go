package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/mail"

	"github.com/Infisical/agent-vault/internal/crypto"
)

// sentinel is the known plaintext encrypted during setup and verified during unlock.
const sentinel = "agent-vault-master-key-check"

// ErrWrongPassword is returned when the master password does not match.
var ErrWrongPassword = errors.New("wrong master password")

// MasterKey holds a derived encryption key in memory.
type MasterKey struct {
	key []byte
}

// Key returns the raw 32-byte encryption key.
func (mk *MasterKey) Key() []byte {
	return mk.key
}

// Wipe zeros the key material. Call this when the server shuts down.
func (mk *MasterKey) Wipe() {
	crypto.WipeBytes(mk.key)
}

// VerificationRecord holds the artifacts needed to verify a master password
// on subsequent startups.
type VerificationRecord struct {
	Salt   []byte
	Sentinel []byte // encrypted sentinel ciphertext
	Nonce  []byte
	Params crypto.KDFParams
}

// Setup creates a new master key from the given password. It generates a random
// salt, derives an encryption key via Argon2id, and encrypts a sentinel value.
// Returns the master key (for immediate use) and a verification record (to persist).
func Setup(password []byte) (*MasterKey, *VerificationRecord, error) {
	params := crypto.DefaultKDFParams()

	salt, err := crypto.GenerateSalt(int(params.SaltLen))
	if err != nil {
		return nil, nil, fmt.Errorf("generating salt: %w", err)
	}

	key := crypto.DeriveKey(password, salt, params)

	ciphertext, nonce, err := crypto.Encrypt([]byte(sentinel), key)
	if err != nil {
		crypto.WipeBytes(key)
		return nil, nil, fmt.Errorf("encrypting sentinel: %w", err)
	}

	return &MasterKey{key: key}, &VerificationRecord{
		Salt:     salt,
		Sentinel: ciphertext,
		Nonce:    nonce,
		Params:   params,
	}, nil
}

// Unlock derives the encryption key from the password and stored verification
// record, then verifies correctness by decrypting the sentinel. Returns
// ErrWrongPassword if the password is incorrect.
func Unlock(password []byte, record *VerificationRecord) (*MasterKey, error) {
	key := crypto.DeriveKey(password, record.Salt, record.Params)

	plaintext, err := crypto.Decrypt(record.Sentinel, record.Nonce, key)
	if err != nil {
		crypto.WipeBytes(key)
		return nil, ErrWrongPassword
	}

	if string(plaintext) != sentinel {
		crypto.WipeBytes(key)
		return nil, ErrWrongPassword
	}

	return &MasterKey{key: key}, nil
}

// HashUserPassword hashes a user password with a random salt using Argon2id.
// Returns the hash, salt, and the KDF parameters used (for storage alongside the hash).
func HashUserPassword(password []byte) (hash, salt []byte, params crypto.KDFParams, err error) {
	params = crypto.DefaultKDFParams()
	salt, err = crypto.GenerateSalt(int(params.SaltLen))
	if err != nil {
		return nil, nil, params, fmt.Errorf("generating salt: %w", err)
	}
	hash = crypto.DeriveKey(password, salt, params)
	return hash, salt, params, nil
}

// VerifyUserPassword checks a password against a stored hash, salt, and KDF params.
func VerifyUserPassword(password, hash, salt []byte, params crypto.KDFParams) bool {
	derived := crypto.DeriveKey(password, salt, params)
	return subtle.ConstantTimeCompare(derived, hash) == 1
}

// ValidateEmail performs basic email format validation.
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	return nil
}
