package bank

import (
	"../../storage/txpool"
)

type BankAdapter struct {
	Gateway *BankGateway
}

func NewBankAdapter(gateway *BankGateway) *BankAdapter {
	return &BankAdapter{
		Gateway: gateway,
	}
}

func (a *BankAdapter) HandleDeposit(tx *txpool.Transaction) (string, error) {
	return a.Gateway.Transfer("bank-reserve", tx.To, tx.Amount)
}

func (a *BankAdapter) HandleWithdraw(tx *txpool.Transaction) (string, error) {
	return a.Gateway.Transfer(tx.From, "bank-reserve", tx.Amount)
}
