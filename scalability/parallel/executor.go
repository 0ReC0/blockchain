package parallel

import (
	"sync"

	"../../storage/blockchain"
	"../../storage/txpool"
)

type ParallelExecutor struct {
	Workers int
}

func (e *ParallelExecutor) ExecuteTransactions(transactions []*txpool.Transaction, chain *blockchain.Blockchain) error {
	ch := make(chan *txpool.Transaction, len(transactions))
	var wg sync.WaitGroup

	for i := 0; i < e.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tx := range ch {
				// Выполняем транзакцию
				chain.Execute(tx)
			}
		}()
	}

	for _, tx := range transactions {
		ch <- tx
	}
	close(ch)
	wg.Wait()
	return nil
}
