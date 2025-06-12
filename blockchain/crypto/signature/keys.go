// crypto/signature/keys.go

package signature

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// LoadPublicKey загружает публичный ключ из его строкового представления (PEM)
func LoadPublicKey(data string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}

	return ecdsaPubKey, nil
}
