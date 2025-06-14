package signature

import (
	"crypto/ecdsa"
	"fmt"
)

var PublicKeyRegistry = make(map[string]*ecdsa.PublicKey)

func RegisterPublicKey(id string, pubKey *ecdsa.PublicKey) {
	PublicKeyRegistry[id] = pubKey
}

func GetPublicKey(id string) (*ecdsa.PublicKey, error) {
	pubKey, exists := PublicKeyRegistry[id]
	if !exists {
		return nil, fmt.Errorf("public key not found for validator %s", id)
	}
	return pubKey, nil
}
