package hash

import "crypto/sha256"

type SHA256Hasher struct{}

func (h *SHA256Hasher) Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (h *SHA256Hasher) Name() string {
	return "SHA-256"
}
