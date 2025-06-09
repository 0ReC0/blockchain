package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	txpool "../txpool"
)

type Block struct {
	Index        int64
	Timestamp    int64
	PrevHash     string
	Hash         string
	Transactions []txpool.Transaction
	Validator    string
	Nonce        string
}

func NewBlock(index int64, prevHash string, transactions []txpool.Transaction, validator string) *Block {
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
	record := string(b.Index) + string(b.Timestamp) + b.PrevHash + b.Validator
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
