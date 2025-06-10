package txpool

import (
	"encoding/json"
	"math/big"
	"time"

	"../../crypto/signature"
	"../../privacy/zkp"
)

type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    float64
	Timestamp int64
	Signature string
	IsPrivate bool
	Encrypted []byte
	PublicKey []byte
}

func NewTransaction(from, to string, amount float64) *Transaction {
	return &Transaction{
		ID:        GenerateID(),
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
}

func (t *Transaction) Serialize() []byte {
	data, _ := json.Marshal(t)
	return data
}
func (t *Transaction) Verify() bool {
	pubKey, err := signature.LoadPublicKey(t.From)
	if err != nil {
		return false
	}

	// Используем X-координату как "публичный ключ" для ZKP
	publicKeyForZKP := pubKey.X

	zkp := zkp.NewZKPSystem()
	proof, _ := zkp.ProveKnowledge(big.NewInt(12345)) // здесь должен быть секретный ключ

	if !zkp.VerifyProof(publicKeyForZKP, proof) {
		return false
	}

	// Проверяем подпись транзакции
	return signature.Verify(pubKey, t.Serialize(), []byte(t.Signature))
}
