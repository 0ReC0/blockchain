package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"

	"blockchain/storage/txpool"
)

type Block struct {
	Index        int64                 `json:"index"`
	Timestamp    int64                 `json:"timestamp"`
	PrevHash     string                `json:"prev_hash"`
	Hash         string                `json:"hash"`
	Transactions []*txpool.Transaction `json:"transactions"`
	Validator    string                `json:"validator"`
	Nonce        string                `json:"nonce"`
	Signature    []byte                `json:"signature"`
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
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(b); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (b *Block) Deserialize(data []byte) error {
	type BlockTemp struct {
		Index        int64
		Timestamp    int64
		PrevHash     string
		Hash         string
		Transactions []*txpool.Transaction
		Validator    string
		Nonce        string
	}

	var temp BlockTemp
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&temp); err != nil {
		return err
	}

	b.Index = temp.Index
	b.Timestamp = temp.Timestamp
	b.PrevHash = temp.PrevHash
	b.Hash = temp.Hash
	b.Transactions = temp.Transactions
	b.Validator = temp.Validator
	b.Nonce = temp.Nonce

	return nil
}

type BlockWithoutSignature struct {
	Index        int64
	Timestamp    int64
	PrevHash     string
	Hash         string
	Transactions []*txpool.Transaction
	Validator    string
	Nonce        string
}

func (b *Block) SerializeWithoutSignature() []byte {
	type BlockTemp struct {
		Index        int64
		Timestamp    int64
		PrevHash     string
		Hash         string
		Transactions []*txpool.Transaction
		Validator    string
		Nonce        string
	}

	temp := BlockTemp{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		PrevHash:     b.PrevHash,
		Hash:         b.Hash,
		Transactions: b.Transactions,
		Validator:    b.Validator,
		Nonce:        b.Nonce,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(temp); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
