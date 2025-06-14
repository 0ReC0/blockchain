package parallel

import (
	"sync"

	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

type ParallelExecutor struct {
	Workers   int
	BlockSize int // максимальное количество транзакций в блоке

}

func NewParallelExecutor(workers, blockSize int) *ParallelExecutor {
	return &ParallelExecutor{
		Workers:   workers,
		BlockSize: blockSize,
	}
}

func (e *ParallelExecutor) ExecuteTransactions(transactions []*txpool.Transaction, chain *blockchain.Blockchain, validator string) error {
	ch := make(chan []*txpool.Transaction)
	var wg sync.WaitGroup

	// Запускаем воркеров
	for i := 0; i < e.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range ch {
				// Создаём блок
				prevBlock := chain.GetLatestBlock()
				block := blockchain.NewBlock(
					prevBlock.Index+1,
					prevBlock.Hash,
					batch,
					validator,
				)

				// Добавляем блок в цепочку
				chain.AddBlock(block)
			}
		}()
	}

	// Разбиваем транзакции на пакеты
	for i := 0; i < len(transactions); i += e.BlockSize {
		end := i + e.BlockSize
		if end > len(transactions) {
			end = len(transactions)
		}
		ch <- transactions[i:end]
	}
	close(ch)

	wg.Wait()
	return nil
}
