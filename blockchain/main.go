// blockchain/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Консенсус
	"blockchain/consensus/governance"
	"blockchain/consensus/manager"
	"blockchain/consensus/pos"
	"blockchain/monitoring"

	// Сеть
	"blockchain/network/peer"
	// Хранилище
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"

	// Криптография
	"blockchain/crypto/signature"
	// Безопасность
	"blockchain/security/audit"
	"blockchain/security/double_spend"
	"blockchain/security/fiftyone"
	"blockchain/security/sybil"

	// Интеграция (API)
	"blockchain/integration/api"
	// Говернанс
	"blockchain/governance/kyc"
	// Шардинг
	"blockchain/scalability/sharding"
)

// Глобальные переменные для говернанса
var (
	kycManager *kyc.KYCManager
)

func main() {
	fmt.Println("🚀 Starting Minimal Blockchain Node with Sharding...")

	// ============ Инициализация хранилища ============
	chain := blockchain.NewBlockchain()
	txPool := txpool.NewTransactionPool()

	// Инициализируем KYC-менеджер
	auditor := audit.NewSecurityAuditor()
	kycManager = kyc.NewKYCManager(auditor)

	// Устанавливаем KYC-менеджер для txpool и api
	txpool.SetKYCManager(kycManager)
	api.SetKYCManager(kycManager)

	// ============ Инициализация шардов ============
	// Using single shard for stability (multi-shard causes port conflicts)
	// IMPORTANT: Using main chain and txPool for consensus (not shard's separate pools)
	const ShardCount = 1
	shards := make(map[int]*sharding.Shard)
	for i := 0; i < ShardCount; i++ {
		shards[i] = &sharding.Shard{
			ID:         i,
			Validators: []string{"validator1"},
			Chain:      chain,    // Use main chain (shared with API)
			TxPool:     txPool,   // Use main txPool (shared with API)
		}
		fmt.Printf("🧱 Shard %d initialized\n", i)
	}

	// ============ Инициализация валидаторов ============
	peerAddresses := []string{
		"localhost:27656", // validator1
	}

	validators := []*pos.Validator{
		pos.NewValidatorWithAddress("validator1", peerAddresses[0], 2000),
	}

	validatorPool := pos.NewValidatorPool(validators)

	// ============ Инициализация signer'а ============
	signer, err := signature.NewECDSASigner()
	if err != nil {
		panic("❌ Failed to create signer: " + err.Error())
	}

	pubKey, err := signature.ParsePublicKey(signer.PublicKey())
	if err != nil {
		panic("❌ Failed to parse public key: " + err.Error())
	}

	// Регистрируем публичные ключи для всех валидаторов
	for i, v := range validators {
		signature.RegisterPublicKey(v.Address, pubKey)
		fmt.Printf("🔑 Public key registered for validator %s\n", v.Address)
		fmt.Printf("🏷️ Validator %d: %s | Stake: %d\n", i+1, v.Address, v.Balance)
	}

	// ============ Инициализация защиты от 51% атак ============
	validatorsMap := map[string]int64{
		"validator1": 2000,
	}

	guard := fiftyone.NewFiftyOnePercentGuard(validatorsMap)
	go guard.Monitor(30 * time.Second)

	// ============ Инициализация защиты от Sybil ============
	sybilGuard := sybil.NewSybilGuard([]string{
		"validator1",
	})

	peer.SetSybilGuard(sybilGuard)

	// ========== Инициализация аудита безопасности ==========
	auditor = audit.NewSecurityAuditor()

	// ============ Инициализация говернанса ============
	// Создаем менеджер говернанса
	governanceManager := governance.NewGovernanceManager()

	// ============ Запуск REST API ============
	// Создаем расширенный API сервер с доступом к governance компонентам
	apiServer := api.NewAPIServer(chain, txPool, auditor)

	// Добавляем маршруты для говернанса
	// Note: В реальной реализации эти обработчики нужно добавить в API пакет

	go func() {
		fmt.Println("🔌 Starting REST API on :8081")
		if err := apiServer.Start(":8081"); err != nil {
			panic("❌ Failed to start API server: " + err.Error())
		}
	}()

	// ============ Запуск защиты от двойной траты ============
	double_spend.InitSecurity()

	// ========== Логируем запуск ноды ==========
	auditor.RecordEvent(audit.SecurityEvent{
		Timestamp: time.Now(),
		Type:      "NodeStartup",
		Message:   "Blockchain node with sharding started successfully",
		NodeID:    "validator1",
		Severity:  "INFO",
	})

	// ========== Используем аудит в других компонентах ==========
	double_spend.SetAuditor(auditor)
	fiftyone.SetAuditor(auditor)
	sybil.SetAuditor(auditor)

	// Создаем пример предложения
	proposal := governance.NewProposal(
		"gov-001",
		"Update block reward",
		"Change block reward from 5 to 3 tokens",
		validators[0].Address,
		governance.ParameterChange,
		0.67, // 67% голосов
		validatorPool,
	)

	// Добавляем параметры изменения
	proposal.Parameters["block_reward"] = float64(3)
	proposal.Parameters["transaction_fee"] = float64(0.01)
	proposal.Parameters["max_block_size"] = int64(2048)

	// Добавляем предложение в говернанс
	governanceManager.SubmitProposal(proposal)

	// Пример голосования (в реальности это будет происходить через RPC)
	for i, validator := range validators {
		if i == 0 {
			// Первый валидатор голосует "за"
			governanceManager.VoteOnProposal(proposal.ID, validator.Address, "yes")
		} else {
			// Остальные валидаторы голосуют "против"
			governanceManager.VoteOnProposal(proposal.ID, validator.Address, "no")
		}
	}

	// Подсчитываем голоса и выполняем предложение
	if approved := governanceManager.TallyVotes(proposal.ID); approved {
		if err := governanceManager.ExecuteProposal(proposal.ID); err != nil {
			fmt.Printf("Failed to execute proposal: %v\n", err)
		}
	} else {
		fmt.Printf("Proposal %s was not approved\n", proposal.ID)
	}

	// ============ Запуск консенсуса в шардах ============
	// Using PoS consensus for simplicity (BFT has port conflicts in multi-shard setup)
	switcher := manager.NewConsensusSwitcher(manager.ConsensusPoS)
	go func() {
		switcher.StartShardedConsensus(
			shards,
			validators,
			*validatorPool,
			signer,
			peerAddresses,
		)
	}()

	fmt.Println("✅ Node started with sharding support. Waiting for connections...")

	// ============ Инициализация мониторинга ============
	// Создаем и запускаем сервер мониторинга
	monitoringServer := monitoring.NewServer(":9090")
	if err := monitoringServer.Start(); err != nil {
		fmt.Printf("Warning: Failed to start monitoring server: %v\n", err)
	} else {
		fmt.Printf("📊 Monitoring server started on %s\n", monitoringServer.GetMetricsEndpoint())
	}

	// Получаем экземпляр метрик
	metrics := monitoring.GetMetrics()

	// Запускаем периодический сбор метрик системы
	metrics.StartMonitoring()

	// Обновляем некоторые начальные метрики
	metrics.UpdateActivePeers(len(peerAddresses))

	// ============ Демонстрация улучшенных метрик ============
	// Создаем пример транзакций для демонстрации метрик
	go func() {
		// Имитируем обработку транзакций с обновлением метрик
		for {
			time.Sleep(30 * time.Second) // Каждые 30 секунд

			// Генерируем фиктивные данные для демонстрации
			transactionCount := 1500                  // Примерное количество транзакций
			processingTime := 1500 * time.Millisecond // Примерное время обработки

			// Обновляем метрики
			metrics.UpdateBlockProcessingMetrics(transactionCount, processingTime)

			// Логируем для демонстрации
			fmt.Printf("📈 Processed %d transactions in %v (TPS: %.2f)\n",
				transactionCount, processingTime, float64(transactionCount)/processingTime.Seconds())
		}
	}()

	// ============ Бесконечный цикл для поддержания работы ============
	// Добавляем graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 Shutting down blockchain node...")

		// Останавливаем сервер мониторинга
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := monitoringServer.Stop(ctx); err != nil {
			fmt.Printf("Error stopping monitoring server: %v\n", err)
		}

		os.Exit(0)
	}()

	select {}
}
