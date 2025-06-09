package crosschain

import (
	"../../storage/blockchain"
	"../../storage/txpool"
)

type ChainBridge struct {
	Source *blockchain.Blockchain
	Target *blockchain.Blockchain
}

func NewChainBridge(source, target *blockchain.Blockchain) *ChainBridge {
	return &ChainBridge{
		Source: source,
		Target: target,
	}
}

func (b *ChainBridge) LockTokens(addr string, amount float64) *txpool.Transaction {
	tx := txpool.NewTransaction(addr, "bridge-lock", amount)
	b.Source.Transactions = append(b.Source.Transactions, tx)
	return tx
}

func (b *ChainBridge) MintTokens(tx *txpool.Transaction) {
	newTx := &txpool.Transaction{
		ID:        tx.ID,
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Timestamp: tx.Timestamp,
	}
	b.Target.Transactions = append(b.Target.Transactions, newTx)
}
