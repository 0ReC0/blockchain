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
	if _, exists := p.Transactions[tx.ID]; !exists {
		p.Transactions[tx.ID] = tx
	}
}
func (p *TransactionPool) GetTransactions(limit int) []*Transaction {
	p.mu.Lock()
	defer p.mu.Unlock()
	var list []*Transaction
	for id, tx := range p.Transactions {
		list = append(list, tx)
		delete(p.Transactions, id) // Удаляем сразу при взятии
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
