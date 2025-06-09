package private_tx

import (
	"crypto/sha256"
)

func GenerateKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}
