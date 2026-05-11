package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"blockchain/storage/txpool"
)

const API_BASE = "http://localhost:8081"

// Transaction - структура для отправки через API
type Transaction struct {
	ID        string  `json:"id"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Fee       float64 `json:"fee"`
	Timestamp int64   `json:"timestamp"`
	Signature string  `json:"signature"`
}

// TestAPI_SendTransactions - тест отправки транзакций через REST API
func TestAPI_SendTransactions(t *testing.T) {
	numTxs := 1000
	successCount := 0
	var mu sync.Mutex
	
	startTime := time.Now()
	
	var wg sync.WaitGroup
	for i := 0; i < numTxs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			tx := Transaction{
				ID:        fmt.Sprintf("api-tx-%d", idx),
				From:      "sender-api",
				To:        "receiver-api",
				Amount:    100.0,
				Fee:       0.01,
				Timestamp: time.Now().UnixNano(),
				Signature: "test-sig",
			}
			
			jsonData, err := json.Marshal(tx)
			if err != nil {
				t.Errorf("Failed to marshal tx %d: %v", idx, err)
				return
			}
			
			resp, err := http.Post(API_BASE+"/transactions", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Errorf("Failed to send tx %d: %v", idx, err)
				return
			}
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	elapsed := time.Since(startTime)
	tps := float64(successCount) / elapsed.Seconds()
	
	fmt.Printf("\n=== API ТЕСТ: ОТПРАВКА ТРАНЗАКЦИЙ ===\n")
	fmt.Printf("Отправлено транзакций: %d\n", numTxs)
	fmt.Printf("Успешно: %d\n", successCount)
	fmt.Printf("Время выполнения: %v\n", elapsed)
	fmt.Printf("Пропускная способность: %.2f TPS\n", tps)
	fmt.Printf("======================================\n")
}

// TestAPI_GetTransactions - тест получения транзакций через API
func TestAPI_GetTransactions(t *testing.T) {
	resp, err := http.Get(API_BASE + "/transactions")
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	
	var txs []txpool.Transaction
	if err := json.Unmarshal(body, &txs); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	fmt.Printf("\n=== API ТЕСТ: ПОЛУЧЕНИЕ ТРАНЗАКЦИЙ ===\n")
	fmt.Printf("Найдено транзакций в пуле: %d\n", len(txs))
	fmt.Printf("=======================================\n")
}

// TestAPI_GetBlocks - тест получения блоков через API
func TestAPI_GetBlocks(t *testing.T) {
	resp, err := http.Get(API_BASE + "/blocks")
	if err != nil {
		t.Fatalf("Failed to get blocks: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	
	fmt.Printf("\n=== API ТЕСТ: ПОЛУЧЕНИЕ БЛОКОВ ===\n")
	fmt.Printf("Ответ API: %s\n", string(body))
	fmt.Printf("==================================\n")
}

// BenchmarkAPI_SendTransaction - бенчмарк отправки одной транзакции
func BenchmarkAPI_SendTransaction(b *testing.B) {
	tx := Transaction{
		ID:        "bench-tx",
		From:      "sender",
		To:        "receiver",
		Amount:    100.0,
		Fee:       0.01,
		Timestamp: time.Now().UnixNano(),
		Signature: "test-sig",
	}
	
	jsonData, _ := json.Marshal(tx)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(API_BASE+"/transactions", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
	
	b.ReportMetric(float64(b.N)/float64(b.Elapsed().Seconds()), "tx/sec")
}

// BenchmarkAPI_ParallelTransactions - бенчмарк параллельной отправки
func BenchmarkAPI_ParallelTransactions(b *testing.B) {
	numWorkers := 8
	txsPerWorker := b.N / numWorkers
	
	jsonData, _ := json.Marshal(Transaction{
		ID:        "bench-tx",
		From:      "sender",
		To:        "receiver",
		Amount:    100.0,
		Fee:       0.01,
		Timestamp: time.Now().UnixNano(),
		Signature: "test-sig",
	})
	
	b.ResetTimer()
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < txsPerWorker; i++ {
				resp, err := http.Post(API_BASE+"/transactions", "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					b.Error(err)
					continue
				}
				resp.Body.Close()
			}
		}(w)
	}
	wg.Wait()
	
	b.ReportMetric(float64(b.N)/float64(b.Elapsed().Seconds()), "tx_parallel/sec")
}
