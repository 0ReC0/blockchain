package hash

import (
	"golang.org/x/crypto/sha3"
)

type Keccak256Hasher struct{}

func (h *Keccak256Hasher) Hash(data []byte) []byte {
	hash := sha3.Sum256(data)
	return hash[:]
}

func (h *Keccak256Hasher) Name() string {
	return "Keccak-256"
}
