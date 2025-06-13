// crypto/signature/keys.go

package signature

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
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

func LoadPublicKeyFromFile(path string) (*ecdsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Декодируем PEM-данные
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block as certificate")
	}

	// Парсим сертификат
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Извлекаем публичный ключ и проверяем его тип
	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not ECDSA")
	}

	return pubKey, nil
}
