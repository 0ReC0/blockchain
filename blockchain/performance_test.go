package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

// BenchmarkTransactionGeneration - тест генерации транзакций (имитационный режим)
func BenchmarkTransactionGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &txpool.Transaction{
			ID:        fmt.Sprintf("tx-%d", i),
			From:      "sender",
			To:        "receiver",
			Amount:    100.0,
			Timestamp: time.Now().UnixNano(),
		}
	}
	
	b.ReportMetric(float64(b.N)/float64(b.Elapsed().Seconds()), "tx_gen/sec")
}

// BenchmarkParallelTransactionGeneration - тест параллельной генерации (8 шардов)
func BenchmarkParallelTransactionGeneration(b *testing.B) {
	numWorkers := 8
	txsPerWorker := b.N / numWorkers
	
	b.ResetTimer()
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < txsPerWorker; i++ {
				_ = &txpool.Transaction{
					ID:        fmt.Sprintf("tx-%d-%d", workerID, i),
					From:      "sender",
					To:        "receiver",
					Amount:    100.0,
					Timestamp: time.Now().UnixNano(),
				}
			}
		}(w)
	}
	wg.Wait()
	
	b.ReportMetric(float64(b.N)/float64(b.Elapsed().Seconds()), "tx_gen_parallel/sec")
}

// BenchmarkBlockCreation - тест скорости создания блоков
func BenchmarkBlockCreation(b *testing.B) {
	txs := []*txpool.Transaction{
		{
			ID:        "tx-1",
			From:      "sender",
			To:        "receiver",
			Amount:    100.0,
			Timestamp: time.Now().UnixNano(),
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = blockchain.NewBlock(int64(i), "prev-hash", txs, "validator")
	}
	
	b.ReportMetric(float64(b.N)/float64(b.Elapsed().Seconds()), "blocks/sec")
}

// TestRealTPS - интеграционный тест реальной производительности
func TestRealTPS(t *testing.T) {
	pool := txpool.NewTransactionPool()

	// Создаём тестовые транзакции
	numTxs := 10000
	startTime := time.Now()

	for i := 0; i < numTxs; i++ {
		tx := &txpool.Transaction{
			ID:        fmt.Sprintf("tx-%d", i),
			From:      "sender",
			To:        "receiver",
			Amount:    100.0,
			Timestamp: time.Now().UnixNano(),
		}
		pool.AddTransaction(tx)
	}

	elapsed := time.Since(startTime)
	tps := float64(numTxs) / elapsed.Seconds()

	fmt.Printf("\n=== РЕАЛЬНЫЕ РЕЗУЛЬТАТЫ TPS ===\n")
	fmt.Printf("Количество транзакций: %d\n", numTxs)
	fmt.Printf("Время выполнения: %v\n", elapsed)
	fmt.Printf("Пропускная способность: %.2f TPS\n", tps)
	fmt.Printf("================================\n")
}

// TestBlockCreationTime - тест времени создания блока
func TestBlockCreationTime(t *testing.T) {
	chain := blockchain.NewBlockchain()
	
	testCases := []int{10, 50, 100, 200}
	
	fmt.Printf("\n=== ВРЕМЯ СОЗДАНИЯ БЛОКА ===\n")
	
	for _, numTxs := range testCases {
		txs := make([]*txpool.Transaction, numTxs)
		for i := 0; i < numTxs; i++ {
			txs[i] = &txpool.Transaction{
				ID:        fmt.Sprintf("tx-%d", i),
				From:      "sender",
				To:        "receiver",
				Amount:    100.0,
				Timestamp: time.Now().UnixNano(),
			}
		}
		
		startTime := time.Now()
		block := blockchain.NewBlock(int64(numTxs), "prev-hash", txs, "validator")
		chain.AddBlock(block)
		elapsed := time.Since(startTime)
		
		fmt.Printf("Блок с %3d транзакциями: %12v\n", numTxs, elapsed)
	}
	
	fmt.Printf("==============================\n")
}
