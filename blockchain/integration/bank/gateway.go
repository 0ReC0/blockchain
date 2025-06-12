package bank

import (
	"fmt"
	"time"
)

type BankGateway struct {
	APIKey   string
	Endpoint string
}

func NewBankGateway(apiKey, endpoint string) *BankGateway {
	return &BankGateway{
		APIKey:   apiKey,
		Endpoint: endpoint,
	}
}

func (g *BankGateway) Transfer(accountFrom, accountTo string, amount float64) (string, error) {
	// Здесь будет вызов внешнего API
	fmt.Printf("Transferring %.2f from %s to %s\n", amount, accountFrom, accountTo)
	return fmt.Sprintf("tx-%d", time.Now().UnixNano()), nil
}

func (g *BankGateway) GetBalance(account string) (float64, error) {
	// Здесь будет вызов внешнего API
	return 1000000.00, nil
}
