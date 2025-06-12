package signature

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
)

// Verify проверяет подпись транзакции с помощью публичного ключа
func Verify(pub *ecdsa.PublicKey, data, sig []byte) bool {
	hash := sha256.Sum256(data)

	// Десериализуем DER
	var rs struct {
		R, S *big.Int
	}
	if _, err := asn1.Unmarshal(sig, &rs); err != nil {
		fmt.Printf("❌ Failed to parse DER signature: %v\n", err)
		return false
	}

	// Проверяем подпись
	return ecdsa.Verify(pub, hash[:], rs.R, rs.S)
}
