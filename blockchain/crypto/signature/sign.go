// crypto/signature/sign.go

package signature

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
)

// Sign подписывает данные приватным ключом
// Sign — подписывает данные и возвращает подпись в DER-формате
func Sign(priv *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, err
	}

	// Кодируем подпись в DER-формате
	return asn1.Marshal(ecdsaSignature{R: r, S: s})
}
