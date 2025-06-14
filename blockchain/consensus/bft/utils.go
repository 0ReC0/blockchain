package bft

import (
	"blockchain/consensus/pos"
	"blockchain/crypto/signature"
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
)

func HasQuorum(votes map[string][]byte, validators []*pos.Validator, round, height int64, blockHash []byte) bool {
	totalPower := 0.0
	for _, v := range validators {
		totalPower += float64(v.Balance)
	}
	validVotes := 0.0

	data := []byte(fmt.Sprintf("prevote:%d:%d:%x", height, round, blockHash))
	hash := sha256.Sum256(data)

	for from, sig := range votes {
		for _, v := range validators {
			if v.Address == from {
				pubKey, err := signature.GetPublicKey(v.Address)
				if err != nil {
					continue
				}

				// Проверяем подпись
				if ecdsa.VerifyASN1(pubKey, hash[:], sig) {
					validVotes += float64(v.Balance)
				}
				break
			}
		}
	}

	return validVotes > (2.0/3.0)*totalPower
}
