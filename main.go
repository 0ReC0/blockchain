// main.go
package main

import (
	"fmt"
	"time"

	// Уровень консенсуса
	"blockchain/consensus/bft"
	"blockchain/consensus/manager"
	"blockchain/consensus/pos"

	// Сеть

	// Хранилище
	"blockchain/storage/blockchain"
	"blockchain/storage/sharding"
	"blockchain/storage/txpool"

	// Криптография
	"blockchain/crypto/signature"

	// Безопасность
	"blockchain/security/double_spend"

	// Масштабируемость
	"blockchain/scalability/parallel"

	// API
	"blockchain/integration/api"
	"blockchain/integration/bank"

	// Говернанс

	"blockchain/governance/upgrade"
)

func main() {
	fmt.Println("🚀 Starting Blockchain Simulation System...")

	// ============ Инициализация хранилища ============
	chain := blockchain.NewBlockchain()
	txPool := txpool.NewTransactionPool()

	// ============ Инициализация тестовой транзакции ============
	tx1 := txpool.NewTransaction("A", "B", 10)
	txPool.AddTransaction(tx1)

	// ============ Инициализация валидаторов ============
	validators := []*pos.Validator{
		pos.NewValidator("validator1", 2000),
		pos.NewValidator("validator2", 1000),
	}
	validatorPool := pos.NewValidatorPool(validators)

	// ============ Инициализация signer'а ============
	signer, err := signature.NewECDSASigner()
	if err != nil {
		panic("❌ Failed to create signer: " + err.Error())
	}

	// ============ Инициализация BFT-ноды ============
	bftNode := bft.NewBFTNode(
		"validator1",
		validators[0],
		*validatorPool,
		txPool,
		chain,
		signer,
	)

	// ============ Инициализация ConsensusSwitcher ============
	switcher := manager.NewConsensusSwitcher(manager.ConsensusBFT)

	// ============ Запуск консенсуса через ConsensusSwitcher ============
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			<-ticker.C
			switcher.StartConsensus()
		}
	}()

	// ============ Запуск P2P сети ============
	go bft.StartNetwork()

	// ============ Запуск REST API ============
	apiServer := api.NewAPIServer(chain, txPool)
	go func() {
		fmt.Println("🔌 Starting REST API on :8081")
		if err := apiServer.Start(":8081"); err != nil {
			panic("❌ Failed to start API server: " + err.Error())
		}
	}()

	// ============ Запуск защиты от двойных трат ============
	double_spend.InitSecurity()

	// ============ Инициализация говернанса ============
	upgradeMgr := upgrade.NewUpgradeManager("v1.0.0")
	upgradeMgr.ProposeUpgrade("v2.0.0", "Switch to BFT", time.Now().Add(24*time.Hour))
	if err := upgradeMgr.ApproveUpgrade(); err != nil {
		fmt.Println("⚠️ Approval failed:", err)
	}
	if err := upgradeMgr.ApplyUpgrade(); err != nil {
		fmt.Println("⚠️ Upgrade failed:", err)
	}

	// ============ Инициализация шардов ============
	shardMgr := sharding.NewShardManager()
	shardMgr.CreateShard("0")

	// ============ Инициализация масштабируемости ============
	executor := parallel.NewParallelExecutor(4, 10)
	if err := executor.ExecuteTransactions(txPool.GetTransactions(100), chain); err != nil {
		fmt.Println("⚠️ Parallel execution failed:", err)
	}

	// ============ Инициализация банковского шлюза ============
	bankGateway := bank.NewBankGateway("api-key", "https://bank-api.com")
	_, _ = bankGateway.GetBalance("user123")

	// ============ Запуск BFT-узла ============
	go bftNode.Start()

	fmt.Println("✅ Blockchain system started. Waiting for connections...")

	// ============ Бесконечный цикл для поддержания работы сервера ============
	select {}
}
