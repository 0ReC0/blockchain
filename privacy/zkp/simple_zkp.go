package zkp

import (
	"crypto/rand"
	"math/big"
)

// Простая группа по модулю простого числа
const (
	// Пример: группа по модулю простого числа
	p = "23" // Простое число
	g = "5"  // Генератор
)

var (
	P = big.NewInt(0).SetBytes([]byte(p))
	G = big.NewInt(0).SetBytes([]byte(g))
)

// ZKPSystem — система доказательства знания секрета
type ZKPSystem struct {
	GroupOrder *big.Int
	Generator  *big.Int
}

func NewZKPSystem() *ZKPSystem {
	return &ZKPSystem{
		GroupOrder: P,
		Generator:  G,
	}
}

// ProveKnowledge — доказательство знания секрета
func (z *ZKPSystem) ProveKnowledge(secret *big.Int) ([]byte, error) {
	randExp, _ := rand.Int(rand.Reader, z.GroupOrder)
	commitment := new(big.Int).Exp(z.Generator, randExp, z.GroupOrder)
	challenge, _ := rand.Int(rand.Reader, z.GroupOrder)
	response := new(big.Int).Add(randExp, new(big.Int).Mul(challenge, secret))
	return append(commitment.Bytes(), append(challenge.Bytes(), response.Bytes()...)...), nil
}

// VerifyProof — проверка доказательства
func (z *ZKPSystem) VerifyProof(publicKey *big.Int, proof []byte) bool {
	// Разбираем proof
	commit := new(big.Int).SetBytes(proof[:32])
	challenge := new(big.Int).SetBytes(proof[32:64])
	response := new(big.Int).SetBytes(proof[64:])

	left := new(big.Int).Exp(z.Generator, response, z.GroupOrder)
	right := new(big.Int).Mul(commit, new(big.Int).Exp(publicKey, challenge, z.GroupOrder))
	right.Mod(right, z.GroupOrder)

	return left.Cmp(right) == 0
}
