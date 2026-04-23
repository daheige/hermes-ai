package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var aesKey []byte

// AES-GCM加密的AES密钥长度‌可以是 128位、192位或256位‌（即16字节、24字节、32字节）。
// 其中，‌256位密钥‌（AES-256-GCM）因提供最高安全级别，被广泛用于高安全要求场景，如数据全链路保护和大模型API加密通信。
// 密钥长度越长，安全性越高，但加密解密性能相对略低。

// InitAesKey 初始化aes key
func InitAesKey(secret string) {
	hash := sha256.Sum256([]byte(secret))
	aesKey = hash[:]
}

// Encrypt encrypts plaintext using AES-GCM and returns base64-encoded ciphertext.
// Format: base64(nonce || ciphertext || tag)
func Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext produced by Encrypt.
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted tries to decrypt the value and returns true if it appears to be
// a valid ciphertext produced by Encrypt.
func IsEncrypted(value string) bool {
	_, err := Decrypt(value)
	return err == nil
}
