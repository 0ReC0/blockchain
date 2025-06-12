package txpool

import (
	"sync"
	"time"

	"blockchain/security/double_spend"
)

type TransactionPool struct {
	Transactions map[string]*Transaction
	mu           sync.Mutex
}

func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		Transactions: make(map[string]*Transaction),
	}
}

var doubleSpendGuard *double_spend.DoubleSpendGuard

func init() {
	doubleSpendGuard = double_spend.NewDoubleSpendGuard()
	doubleSpendGuard.StartCleanup(5 * time.Minute)
}

func (p *TransactionPool) AddTransaction(tx *Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !doubleSpendGuard.CheckAndMark(tx.ID) {
		// Транзакция является двойной тратой — не добавляем
		return
	}
	p.Transactions[tx.ID] = tx
}
func (p *TransactionPool) GetTransactions(limit int) []*Transaction {
	p.mu.Lock()
	defer p.mu.Unlock()

	var list []*Transaction
	for _, tx := range p.Transactions {
		list = append(list, tx)
		if len(list) >= limit {
			break
		}
	}
	return list
}

func (p *TransactionPool) RemoveTransaction(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.Transactions, id)
}

func (p *TransactionPool) Flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Transactions = make(map[string]*Transaction)
}
