// blockchain/blockchain.go
package blockchain

import (
	"fmt"
	"sync"

	"blockchain/storage/txpool"

	"github.com/dgraph-io/badger/v4"
)

type Blockchain struct {
	db *badger.DB
	mu sync.Mutex
}

func NewBlockchain() *Blockchain {
	opts := badger.DefaultOptions("./data/badgerdb").WithInMemory(false)
	db, err := badger.Open(opts)
	if err != nil {
		panic("❌ Failed to open BadgerDB: " + err.Error())
	}
	fmt.Println("✅ BadgerDB opened successfully")

	// Проверяем, есть ли уже блоки в БД
	hasGenesis := false
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		it.Rewind()
		hasGenesis = it.Valid()
		return nil
	})

	if err != nil {
		panic("❌ Failed to check for existing blocks: " + err.Error())
	}

	bc := &Blockchain{
		db: db,
	}

	// Если genesis-блока нет — создаём его
	if !hasGenesis {
		genesis := NewGenesisBlock()
		fmt.Println("⛏️ Genesis block created:", genesis.Hash)
		bc.AddBlock(genesis)
	}

	return bc
}

func NewGenesisBlock() *Block {
	return NewBlock(0, "0", []*txpool.Transaction{}, "genesis")
}

func (bc *Blockchain) DB() *badger.DB {
	return bc.db
}

func (bc *Blockchain) GetBlockByNumber(blockNumber interface{}) *Block {
	numStr, ok := blockNumber.(string)
	if !ok {
		return nil
	}

	var num int64
	if len(numStr) > 2 && numStr[:2] == "0x" {
		_, err := fmt.Sscanf(numStr, "%x", &num)
		if err != nil {
			return nil
		}
	} else {
		_, err := fmt.Sscanf(numStr, "%d", &num)
		if err != nil {
			return nil
		}
	}

	var block *Block
	err := bc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		i := int64(0)
		for it.Rewind(); it.Valid(); it.Next() {
			if i == num {
				item := it.Item()
				val, _ := item.ValueCopy(nil)
				block = &Block{}
				block.Deserialize(val)
				return nil
			}
			i++
		}
		return nil
	})

	if err != nil {
		fmt.Println("❌ Failed to get block from BadgerDB:", err)
	}

	return block
}

func (bc *Blockchain) GetLatestBlock() *Block {
	var latestBlock *Block
	err := bc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		opts.PrefetchSize = 1
		it := txn.NewIterator(opts)
		defer it.Close()

		if it.Rewind(); it.Valid() {
			item := it.Item()
			val, _ := item.ValueCopy(nil)
			latestBlock = &Block{}
			latestBlock.Deserialize(val)
		}
		return nil
	})
	if err != nil {
		fmt.Println("❌ Failed to get latest block from BadgerDB:", err)
	}
	return latestBlock
}

func (bc *Blockchain) Close() {
	bc.db.Close()
}
func (bc *Blockchain) AddBlock(block *Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Проверяем, существует ли уже блок с таким хэшем
	if bc.HasBlock(block.Hash) {
		fmt.Printf("❌ Block with hash %s already exists\n", block.Hash)
		return
	}

	err := bc.db.Update(func(txn *badger.Txn) error {
		data := block.Serialize()
		return txn.Set([]byte(block.Hash), data)
	})

	if err != nil {
		fmt.Println("❌ Failed to save block to BadgerDB:", err)
	}
}
func (bc *Blockchain) HasBlock(hash string) bool {
	var exists bool
	bc.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(hash))
		exists = err == nil
		return nil
	})
	return exists
}

// HasTransaction проверяет, существует ли транзакция с данным ID в цепочке
func (bc *Blockchain) HasTransaction(txID string) bool {
	var exists bool
	err := bc.db.View(func(txn *badger.Txn) error {
		// Итерируем по всем блокам
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			block := &Block{}
			if err := block.Deserialize(val); err != nil {
				return err
			}

			// Проверяем транзакции в блоке
			for _, tx := range block.Transactions {
				if tx.ID == txID {
					exists = true
					return nil
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ Failed to check transaction existence: %v\n", err)
	}

	return exists
}
