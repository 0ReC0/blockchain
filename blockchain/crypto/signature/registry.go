// crypto/signature/registry.go

package signature

import (
	"crypto/ecdsa"
	"fmt"
	"sync"
)

var (
	PublicKeyRegistry = make(map[string]*ecdsa.PublicKey)
	pubKeyMu          sync.RWMutex
)

// RegisterPublicKey регистрирует публичный ключ валидатора
func RegisterPublicKey(id string, pubKey *ecdsa.PublicKey) {
	pubKeyMu.Lock()
	defer pubKeyMu.Unlock()
	PublicKeyRegistry[id] = pubKey
}

// GetPublicKey возвращает публичный ключ по идентификатору валидатора
func GetPublicKey(id string) (*ecdsa.PublicKey, error) {
	pubKeyMu.RLock()
	defer pubKeyMu.RUnlock()
	pubKey, exists := PublicKeyRegistry[id]
	if !exists {
		return nil, fmt.Errorf("public key not found for validator %s", id)
	}
	return pubKey, nil
}
