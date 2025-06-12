package txpool

import (
	"encoding/json"
	"fmt"
	"time"

	"blockchain/crypto/signature"
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
	temp := struct {
		From      string
		To        string
		Amount    float64
		Timestamp int64
	}{
		From:      t.From,
		To:        t.To,
		Amount:    t.Amount,
		Timestamp: t.Timestamp,
	}
	data, _ := json.Marshal(temp)
	return data
}

func (t *Transaction) Verify() bool {
	// 1. Получаем публичный ключ из реестра
	pubKey, err := signature.GetPublicKey(t.From)
	if err != nil {
		fmt.Printf("Public key not found for %s: %v\n", t.From, err)
		return false
	}

	fmt.Printf("Raw signature (hex): %x\n", []byte(t.Signature))

	// 2. Проверяем подпись
	if !signature.Verify(pubKey, t.Serialize(), []byte(t.Signature)) {
		fmt.Printf("Signature verification failed for transaction %s\n", t.ID)
		return false
	}

	return true
}
