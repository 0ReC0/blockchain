package shielded

import (
	"sync"

	"../private_tx"
)

type ShieldedPool struct {
	Transactions map[string]*private_tx.PrivateTransaction
	mu           sync.Mutex
}

func NewShieldedPool() *ShieldedPool {
	return &ShieldedPool{
		Transactions: make(map[string]*private_tx.PrivateTransaction),
	}
}

func (p *ShieldedPool) AddTransaction(tx *private_tx.PrivateTransaction) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Transactions[tx.Recipient] = tx
}

func (p *ShieldedPool) GetTransactions(limit int) []*private_tx.PrivateTransaction {
	p.mu.Lock()
	defer p.mu.Unlock()

	var list []*private_tx.PrivateTransaction
	for _, tx := range p.Transactions {
		list = append(list, tx)
		if len(list) >= limit {
			break
		}
	}
	return list
}
