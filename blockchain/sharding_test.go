package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"blockchain/security/audit"
	"blockchain/security/double_spend"
	"blockchain/storage/txpool"
)

// InitSecurityForTests - инициализация для тестов
func init() {
	auditor := audit.NewSecurityAuditor()
	double_spend.SetAuditor(auditor)
}

// BenchmarkShardingPerformance - тест производительности шардирования
func BenchmarkShardingPerformance(b *testing.B) {
	shardCounts := []int{1, 2, 4, 8}
	
	for _, numShards := range shardCounts {
		b.Run(fmt.Sprintf("Shards_%d", numShards), func(b *testing.B) {
			// Создаём независимые пулы для каждого шарда
			pools := make([]*txpool.TransactionPool, numShards)
			for i := 0; i < numShards; i++ {
				pools[i] = txpool.NewTransactionPool()
			}
			
			var counter uint64
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Распределяем транзакции по шардам
				shardIdx := i % numShards
				txID := fmt.Sprintf("tx-%d-%d", shardIdx, atomic.AddUint64(&counter, 1))
				tx := &txpool.Transaction{
					ID:        txID,
					From:      "sender",
					To:        "receiver",
					Amount:    100.0,
					Timestamp: time.Now().UnixNano(),
				}
				pools[shardIdx].AddTransaction(tx)
			}
			
			tps := float64(b.N) / b.Elapsed().Seconds()
			b.ReportMetric(tps, "tps")
		})
	}
}

// TestShardingScalability - тест масштабируемости шардирования
func TestShardingScalability(t *testing.T) {
	shardCounts := []int{1, 2, 4, 8}
	results := make(map[int]float64)
	
	fmt.Printf("\n=== ТЕСТ МАСШТАБИРУЕМОСТИ ШАРДИРОВАНИЯ ===\n\n")
	
	for _, numShards := range shardCounts {
		// Создаём независимые пулы для каждого шарда
		pools := make([]*txpool.TransactionPool, numShards)
		for i := 0; i < numShards; i++ {
			pools[i] = txpool.NewTransactionPool()
		}
		
		// Тестируем производительность
		numTxs := 10000
		startTime := time.Now()
		
		var wg sync.WaitGroup
		var counter uint64
		txsPerShard := numTxs / numShards
		
		// Параллельная обработка по шардам
		for shardIdx := 0; shardIdx < numShards; shardIdx++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				for i := 0; i < txsPerShard; i++ {
					txID := fmt.Sprintf("tx-%d-%d", idx, atomic.AddUint64(&counter, 1))
					tx := &txpool.Transaction{
						ID:        txID,
						From:      "sender",
						To:        "receiver",
						Amount:    100.0,
						Timestamp: time.Now().UnixNano(),
					}
					pools[idx].AddTransaction(tx)
				}
			}(shardIdx)
		}
		
		wg.Wait()
		elapsed := time.Since(startTime)
		tps := float64(numTxs) / elapsed.Seconds()
		results[numShards] = tps
		
		fmt.Printf("Шарды: %2d | TPS: %12.2f | Время: %v\n", numShards, tps, elapsed)
	}
	
	fmt.Printf("==========================================\n\n")
	
	// Анализируем масштабирование
	if len(results) >= 2 {
		singleShardTPS := results[1]
		fmt.Printf("АНАЛИЗ МАСШТАБИРУЕМОСТИ (база: 1 шард = %.0f TPS):\n", singleShardTPS)
		fmt.Printf("----------------------------------------------\n")
		for _, shards := range []int{1, 2, 4, 8} {
			tps := results[shards]
			speedup := tps / singleShardTPS
			efficiency := (speedup / float64(shards)) * 100
			fmt.Printf("%d шард(а/ов): %8.0f TPS | ускорение %.2fx | эффективность %.1f%%\n",
				shards, tps, speedup, efficiency)
		}
	}
}

// TestShardingWithContention - тест с конкуренцией за ресурсы
func TestShardingWithContention(t *testing.T) {
	fmt.Printf("\n=== ТЕСТ ШАРДИРОВАНИЯ С КОНКУРЕНЦИЕЙ ===\n\n")
	
	shardCounts := []int{1, 2, 4, 8}
	
	for _, numShards := range shardCounts {
		pools := make([]*txpool.TransactionPool, numShards)
		for i := 0; i < numShards; i++ {
			pools[i] = txpool.NewTransactionPool()
		}
		
		numTxs := 5000
		startTime := time.Now()
		
		var wg sync.WaitGroup
		var counter uint64
		
		// Все потоки пишут в случайные шарды (конкуренция)
		for w := 0; w < numShards; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < numTxs/numShards; i++ {
					shardIdx := int(atomic.AddUint64(&counter, 1)) % numShards
					txID := fmt.Sprintf("tx-c-%d", atomic.AddUint64(&counter, 1))
					tx := &txpool.Transaction{
						ID:        txID,
						From:      "sender",
						To:        "receiver",
						Amount:    100.0,
						Timestamp: time.Now().UnixNano(),
					}
					pools[shardIdx].AddTransaction(tx)
				}
			}()
		}
		
		wg.Wait()
		elapsed := time.Since(startTime)
		tps := float64(numTxs) / elapsed.Seconds()
		
		fmt.Printf("%d шард(а/ов): %8.2f TPS за %v\n", numShards, tps, elapsed)
	}
	
	fmt.Printf("========================================\n")
}

// TestShardingOverhead - тест накладных расходов
func TestShardingOverhead(t *testing.T) {
	fmt.Printf("\n=== ТЕСТ НАКЛАДНЫХ РАСХОДОВ ===\n\n")
	
	// Тест с 8 независимыми шардами
	pools8 := make([]*txpool.TransactionPool, 8)
	for i := 0; i < 8; i++ {
		pools8[i] = txpool.NewTransactionPool()
	}
	
	numTxs := 10000
	startTime := time.Now()
	
	var wg sync.WaitGroup
	var counter uint64
	txsPerShard := numTxs / 8
	
	for shardIdx := 0; shardIdx < 8; shardIdx++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for i := 0; i < txsPerShard; i++ {
				txID := fmt.Sprintf("tx-%d-%d", idx, atomic.AddUint64(&counter, 1))
				tx := &txpool.Transaction{
					ID:        txID,
					From:      "sender",
					To:        "receiver",
					Amount:    100.0,
					Timestamp: time.Now().UnixNano(),
				}
				pools8[idx].AddTransaction(tx)
			}
		}(shardIdx)
	}
	
	wg.Wait()
	elapsed := time.Since(startTime)
	tps := float64(numTxs) / elapsed.Seconds()
	
	fmt.Printf("8 шардов (параллельно): %.2f TPS за %v\n", tps, elapsed)
	fmt.Printf("=====================================\n")
}
