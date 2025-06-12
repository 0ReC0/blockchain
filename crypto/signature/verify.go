package signature

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"math/big"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
)

// Verify проверяет подпись транзакции с помощью публичного ключа
func Verify(pub *ecdsa.PublicKey, data, sig []byte) bool {
	hash := sha256.Sum256(data)

	r, s, err := parseDERSignature(sig)
	if err != nil {
		fmt.Printf("Failed to parse DER signature: %v\n", err)
		return false
	}

	return ecdsa.Verify(pub, hash[:], r, s)
}
func parseDERSignature(sig []byte) (r, s *big.Int, err error) {
	input := cryptobyte.String(sig)
	var inner cryptobyte.String

	// Читаем SEQUENCE
	if !input.ReadASN1(&inner, asn1.SEQUENCE) || !input.Empty() {
		return nil, nil, fmt.Errorf("invalid ASN.1 sequence")
	}

	rInt := new(big.Int)
	sInt := new(big.Int)

	// Читаем r и s
	if !inner.ReadASN1Integer(rInt) || !inner.ReadASN1Integer(sInt) {
		return nil, nil, fmt.Errorf("invalid integers in signature")
	}

	return rInt, sInt, nil
}
