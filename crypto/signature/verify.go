package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
)

// Verify проверяет подпись транзакции с помощью публичного ключа
func Verify(pub *ecdsa.PublicKey, data, sig []byte) bool {
	if pub.Curve == nil {
		pub.Curve = elliptic.P256()
	}

	hash := sha256.Sum256(data)

	r := new(big.Int)
	s := new(big.Int)
	sigLen := len(sig)
	if sigLen == 0 {
		return false
	}
	r.SetBytes(sig[:sigLen/2])
	s.SetBytes(sig[sigLen/2:])

	return ecdsa.Verify(pub, hash[:], r, s)
}
