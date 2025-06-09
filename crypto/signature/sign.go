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

	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])

	return sig, nil
}
