package txpool

import "time"

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
