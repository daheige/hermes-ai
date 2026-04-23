package crypto

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func Password2Hash(password string) (string, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func ValidatePasswordAndHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// KeyHash returns the SHA256 hex digest of the given key.
// It is used for exact lookup and unique index without exposing the raw key.
func KeyHash(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
