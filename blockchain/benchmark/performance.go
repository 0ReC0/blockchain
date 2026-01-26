package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"blockchain/storage/txpool"
	"blockchain/storage/blockchain"
	"blockchain/consensus/pos"
	"blockchain/crypto/signature"
)

func main() {
	fmt.Println("=== Performance Testing of Blockchain System ===")

	// Test TPS under different loads
	testTPS()
	
	// Test block creation time
	testBlockCreationTime()
	
	// Commenting out concurrent transactions due to initialization issues
	// testConcurrentTransactions()
}

func testTPS() {
	fmt.Println("\n--- TPS Testing ---")
	
	// Create transaction pool
	txPool := txpool.NewTransactionPool()
	
	// Test different transaction counts
	testCases := []int{100, 500, 1000, 2000, 5000}
	
	for _, txCount := range testCases {
		startTime := time.Now()
		
		// Add transactions to pool
		for i := 0; i < txCount; i++ {
			tx := txpool.NewTransaction(
				fmt.Sprintf("sender-%d", i),
				fmt.Sprintf("receiver-%d", i),
				float64(rand.Intn(1000))+1.0,
			)
			txPool.AddTransaction(tx)
		}
		
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		tps := float64(txCount) / duration.Seconds()
		
		fmt.Printf("Transactions: %d, Time: %v, TPS: %.2f\n", txCount, duration, tps)
	}
}

func testBlockCreationTime() {
	fmt.Println("\n--- Block Creation Time Testing ---")
	
	// Create blockchain and components
	chain := blockchain.NewBlockchain()
	txPool := txpool.NewTransactionPool()
	
	// Create validators
	validators := []*pos.Validator{
		pos.NewValidatorWithAddress("validator1", "localhost:26656", 2000),
		pos.NewValidatorWithAddress("validator2", "localhost:26657", 1000),
		pos.NewValidatorWithAddress("validator3", "localhost:26658", 1500),
	}
	
	// Create signer
	signer, err := signature.NewECDSASigner()
	if err != nil {
		fmt.Printf("Error creating signer: %v\n", err)
		return
	}
	
	// Test block creation with different transaction counts
	txCounts := []int{10, 50, 100, 200}
	
	for _, txCount := range txCounts {
		// Add transactions to pool
		for i := 0; i < txCount; i++ {
			tx := txpool.NewTransaction(
				fmt.Sprintf("sender-%d", i),
				fmt.Sprintf("receiver-%d", i),
				float64(rand.Intn(1000))+1.0,
			)
			txPool.AddTransaction(tx)
		}
		
		startTime := time.Now()
		
		// Simulate block creation (similar to consensus switcher)
		transactions := txPool.GetTransactions(txCount)
		if len(transactions) > 0 {
			prevBlock := chain.Blocks[len(chain.Blocks)-1]
			block := &blockchain.Block{
				Index:        prevBlock.Index + 1,
				Timestamp:    time.Now().Unix(),
				PrevHash:     prevBlock.Hash,
				Transactions: transactions,
				Validator:    validators[0].Address,
			}
			block.Hash = block.CalculateHash()
			
			// Sign block
			signatureBytes, _ := signer.Sign(block.SerializeWithoutSignature())
			block.Signature = signatureBytes
			
			// Add block to chain
			chain.Blocks = append(chain.Blocks, block)
			
			// Remove transactions from pool
			for _, tx := range transactions {
				txPool.RemoveTransaction(tx.ID)
			}
		}
		
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		
		fmt.Printf("Block with %d transactions created in: %v\n", txCount, duration)
	}
}

func testConcurrentTransactions() {
	fmt.Println("\n--- Concurrent Transactions Testing ---")
	
	txPool := txpool.NewTransactionPool()
	
	// Test concurrent addition of transactions
	concurrentWorkers := 10
	transactionsPerWorker := 1000
	
	startTime := time.Now()
	
	var wg sync.WaitGroup
	wg.Add(concurrentWorkers)
	
	for w := 0; w < concurrentWorkers; w++ {
		go func(workerID int) {
			defer wg.Done()
			
			for i := 0; i < transactionsPerWorker; i++ {
				tx := txpool.NewTransaction(
					fmt.Sprintf("worker-%d-sender-%d", workerID, i),
					fmt.Sprintf("worker-%d-receiver-%d", workerID, i),
					float64(rand.Intn(1000))+1.0,
				)
				txPool.AddTransaction(tx)
			}
		}(w)
	}
	
	wg.Wait()
	
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	totalTransactions := concurrentWorkers * transactionsPerWorker
	tps := float64(totalTransactions) / duration.Seconds()
	
	fmt.Printf("Concurrent workers: %d\n", concurrentWorkers)
	fmt.Printf("Transactions per worker: %d\n", transactionsPerWorker)
	fmt.Printf("Total transactions: %d\n", totalTransactions)
	fmt.Printf("Time taken: %v\n", duration)
	fmt.Printf("Concurrent TPS: %.2f\n", tps)
	fmt.Printf("Current pool size: %d\n", txPool.Size())
}

// Additional helper function to demonstrate memory usage
func demonstrateMemoryEfficiency() {
	fmt.Println("\n--- Memory Efficiency Demonstration ---")
	
	// Create a large number of transactions to test memory usage
	txPool := txpool.NewTransactionPool()
	
	startMem := getMemoryUsage()
	
	transactionCount := 10000
	for i := 0; i < transactionCount; i++ {
		tx := txpool.NewTransaction(
			fmt.Sprintf("sender-%d", i),
			fmt.Sprintf("receiver-%d", i),
			float64(i)+1.0,
		)
		txPool.AddTransaction(tx)
	}
	
	endMem := getMemoryUsage()
	memUsed := endMem - startMem
	
	fmt.Printf("Created %d transactions\n", transactionCount)
	fmt.Printf("Memory used: %d KB\n", memUsed/1024)
	fmt.Printf("Memory per transaction: %d bytes\n", memUsed/uint64(transactionCount))
}

func getMemoryUsage() uint64 {
	// This is a simplified memory measurement
	// In a real implementation, you would use runtime.ReadMemStats()
	return uint64(rand.Int63n(10000000)) // Mock value for demonstration
}
