package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func CalculateHash(data string, key string) string {
	if key == "" {
		return ""
	}

	h := sha256.New()

	h.Write([]byte(data + key))

	hashBytes := h.Sum(nil)

	return hex.EncodeToString(hashBytes)
}

func VerifyHash(data string, key string, hash string) error {
	if key == "" {
		return nil
	}

	calculatedHash := CalculateHash(data, key)
	if calculatedHash != hash {
		return fmt.Errorf("hash verification failed: expected %s, got %s", hash, calculatedHash)
	}

	return nil
}
