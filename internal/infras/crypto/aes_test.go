package crypto

import (
	"fmt"
	"log"
	"testing"
)

func TestAesGCM(t *testing.T) {
	// 32位长度的key
	aesKey = []byte("6b77385e310c709bd86ef23342980d62")
	s := "hello world"
	encrypted, err := Encrypt(s)
	if err != nil {
		t.Errorf("Encrypt failed: %v", err)
	}

	log.Println(encrypted)
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
	}

	log.Println("decrypted:", decrypted)
}

func TestKeyHash(t *testing.T) {
	fmt.Println(KeyHash("1111"))
}
