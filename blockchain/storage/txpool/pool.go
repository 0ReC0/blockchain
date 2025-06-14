package txpool

import (
	"fmt"
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

	// Проверяем, есть ли уже такая транзакция в пуле
	if _, exists := p.Transactions[tx.ID]; exists {
		fmt.Printf("❌ Transaction %s already exists in pool\n", tx.ID)
		return
	}

	// Проверяем, не является ли это двойной тратой
	if !doubleSpendGuard.CheckAndMark(tx.ID) {
		fmt.Printf("❌ Double spend detected for transaction %s\n", tx.ID)
		return
	}

	p.Transactions[tx.ID] = tx
	fmt.Printf("📥 Transaction %s added to pool\n", tx.ID)
}
func (p *TransactionPool) GetTransactions(limit int) []*Transaction {
	p.mu.Lock()
	defer p.mu.Unlock()

	var list []*Transaction
	seen := make(map[string]bool)

	for _, tx := range p.Transactions {
		if seen[tx.ID] {
			continue
		}

		list = append(list, tx)
		seen[tx.ID] = true

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

// HasTransaction проверяет, существует ли транзакция с данным ID в пуле
func (p *TransactionPool) HasTransaction(txID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, exists := p.Transactions[txID]
	return exists
}
