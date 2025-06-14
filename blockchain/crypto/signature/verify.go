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

	// Структура для десериализации DER-подписи
	var rs struct {
		R, S *big.Int
	}

	// Декодируем DER-подпись
	rest, err := asn1.Unmarshal(sig, &rs)
	if err != nil || len(rest) != 0 {
		fmt.Printf("❌ Failed to parse DER signature: %v\n", err)
		return false
	}

	// Проверяем подпись
	return ecdsa.Verify(pub, hash[:], rs.R, rs.S)
}
