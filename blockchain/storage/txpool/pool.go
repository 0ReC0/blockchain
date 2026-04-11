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
	for _, tx := range p.Transactions {
		list = append(list, tx)
		if len(list) >= limit {
			break
		}
	}
	return list
}

// GetAllTransactions returns all transactions without removing them (for API)
func (p *TransactionPool) GetAllTransactions() []*Transaction {
	p.mu.Lock()
	defer p.mu.Unlock()
	var list []*Transaction
	for _, tx := range p.Transactions {
		list = append(list, tx)
	}
	return list
}

func (p *TransactionPool) RemoveTransactions(ids []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, id := range ids {
		delete(p.Transactions, id)
	}
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

func (p *TransactionPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.Transactions)
}
