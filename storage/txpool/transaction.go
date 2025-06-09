package txpool

import (
	"encoding/json"
	"time"

	"../../crypto/signature"
)

type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    float64
	Timestamp int64
	Signature string
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
	return signature.Verify(pubKey, t.Serialize(), []byte(t.Signature))
}
