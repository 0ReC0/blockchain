// crypto/signature/registry.go

package signature

import (
	"crypto/ecdsa"
	"fmt"
)

var PublicKeyRegistry = make(map[string]*ecdsa.PublicKey)

// RegisterPublicKey регистрирует публичный ключ валидатора
func RegisterPublicKey(id string, pubKey *ecdsa.PublicKey) {
	PublicKeyRegistry[id] = pubKey
}

// GetPublicKey возвращает публичный ключ по идентификатору валидатора
func GetPublicKey(id string) (*ecdsa.PublicKey, error) {
	pubKey, exists := PublicKeyRegistry[id]
	if !exists {
		return nil, fmt.Errorf("public key not found for validator %s", id)
	}
	return pubKey, nil
}
