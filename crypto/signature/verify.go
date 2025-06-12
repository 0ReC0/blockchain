package signature

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"math/big"
)

// Verify проверяет подпись транзакции с помощью публичного ключа
func Verify(pub *ecdsa.PublicKey, data, sig []byte) bool {
	hash := sha256.Sum256(data)
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	return ecdsa.Verify(pub, hash[:], r, s)
}
