package txpool

import "sync"

type TransactionPool struct {
	Transactions map[string]*Transaction
	mu           sync.Mutex
}

func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		Transactions: make(map[string]*Transaction),
	}
}

func (p *TransactionPool) AddTransaction(tx *Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()
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
