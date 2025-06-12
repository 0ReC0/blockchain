// crypto/signature/sign.go

package signature

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
)

// Sign подписывает данные приватным ключом
func Sign(priv *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, err
	}

	// Формируем подпись как 64 байта: r (32) + s (32)
	sig := make([]byte, 64)
	copy(sig[:32], r.Bytes())
	copy(sig[32:], s.Bytes())

	return sig, nil
}