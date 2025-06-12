package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"blockchain/storage/txpool"
)

type Block struct {
	Index        int64
	Timestamp    int64
	PrevHash     string
	Hash         string
	Transactions []*txpool.Transaction
	Validator    string
	Nonce        string
	Signature    []byte
}

func NewBlock(index int64, prevHash string, transactions []*txpool.Transaction, validator string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevHash,
		Transactions: transactions,
		Validator:    validator,
	}
	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) CalculateHash() string {
	hashData := fmt.Sprintf("%d:%d:%s:%s:%s:%x",
		b.Index,
		b.Timestamp,
		b.PrevHash,
		b.TransactionsHash(),
		b.Validator,
		b.Nonce,
	)
	hash := sha256.Sum256([]byte(hashData))
	return fmt.Sprintf("%x", hash)
}

func (b *Block) TransactionsHash() string {
	// Упрощённый способ хэширования транзакций
	// В реальной системе используйте Merkle Tree
	var txHashes string
	for _, tx := range b.Transactions {
		txHashes += tx.ID
	}
	hash := sha256.Sum256([]byte(txHashes))
	return fmt.Sprintf("%x", hash)
}

func (b *Block) Serialize() []byte {
	data, _ := json.Marshal(b)
	return data
}

func (b *Block) Deserialize(data []byte) error {
	return json.Unmarshal(data, b)
}
