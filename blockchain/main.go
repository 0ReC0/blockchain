// blockchain/main.go
package main

import (
	"fmt"
	"time"

	// Консенсус
	"blockchain/consensus/manager"
	"blockchain/consensus/pos"
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
	"blockchain/consensus/governance"
	// Шардинг
	"blockchain/scalability/sharding"
)

// Глобальные переменные для говернанса
var (
	kycManager *kyc.KYCManager
)

func generateID() string {
	// Упрощенная реализация генерации ID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func validatorExists(addr string, pool *pos.ValidatorPool) bool {
	for _, v := range *pool {
		if v.Address == addr {
			return true
		}
	}
	return false
}

// ============ Тестовый сценарий для off-chain ============
func runOffChainTestScenario(api *api.OffChainHandler) {
	fmt.Println("🧪 Starting off-chain test scenario...")

	// 1. Создание канала
	channelID := "test-channel-1"
	req := struct {
		ID       string  `json:"id"`
		AddrA    string  `json:"addr_a"`
		AddrB    string  `json:"addr_b"`
		DepositA float64 `json:"deposit_a"`
		DepositB float64 `json:"deposit_b"`
		PubKeyA  string  `json:"pubkey_a"`
		PubKeyB  string  `json:"pubkey_b"`
		Timeout  int64   `json:"timeout"`
	}{
		ID:       channelID,
		AddrA:    "validator1",
		AddrB:    "validator2",
		DepositA: 100,
		DepositB: 50,
		PubKeyA:  "pubkeyA",
		PubKeyB:  "pubkeyB",
		Timeout:  time.Now().Add(24 * time.Hour).Unix(),
	}

	body, _ := json.Marshal(req)
	r := httptest.NewRequest("POST", "/offchain/channel/create", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	api.HandleCreateChannel(w, r)
	fmt.Println("✅ Channel created:", channelID)

	// 2. Обновление состояния (оффчейн транзакции)
	for i := 0; i < 3; i++ {
		reqUpdate := struct {
			ID      string  `json:"id"`
			AmountA float64 `json:"amount_a"`
			AmountB float64 `json:"amount_b"`
			Nonce   int     `json:"nonce"`
			SigA    string  `json:"sig_a"`
			SigB    string  `json:"sig_b"`
		}{
			ID:      channelID,
			AmountA: 90 - float64(i)*5,
			AmountB: 60 + float64(i)*5,
			Nonce:   i + 1,
			SigA:    "sigA",
			SigB:    "sigB",
		}
		body, _ = json.Marshal(reqUpdate)
		r = httptest.NewRequest("POST", "/offchain/channel/update", bytes.NewBuffer(body))
		w = httptest.NewRecorder()
		api.HandleUpdateChannel(w, r)
		fmt.Printf("✅ Channel updated: %s (Nonce: %d)\n", channelID, i+1)
	}

	// 3. Финализация
	reqFinal := struct {
		ID string `json:"id"`
	}{
		ID: channelID,
	}
	body, _ = json.Marshal(reqFinal)
	r = httptest.NewRequest("POST", "/offchain/channel/finalize", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	api.HandleFinalizeChannel(w, r)
	fmt.Println("✅ Channel finalized:", channelID)
}

func main() {
	fmt.Println("🚀 Starting Minimal Blockchain Node with Sharding...")

	// ============ Инициализация хранилища ============
	chain := blockchain.NewBlockchain()
	txPool := txpool.NewTransactionPool()

	// Инициализируем KYC-менеджер
	auditor := audit.NewSecurityAuditor()
	kycManager = kyc.NewKYCManager(auditor)

	// ============ Инициализация шардов ============
	const ShardCount = 4
	shards := make(map[int]*sharding.Shard)
	for i := 0; i < ShardCount; i++ {
		shards[i] = &sharding.Shard{
			ID:        i,
			Validators: []string{"validator1", "validator2", "validator3"},
			Chain:     blockchain.NewBlockchain(),
			TxPool:    txpool.NewTransactionPool(),
		}
		fmt.Printf("🧱 Shard %d initialized\n", i)
	}

	// Роутер для маршрутизации транзакций
	router := &sharding.ShardRouter{ShardCount: len(shards)}
	shardingManager := &sharding.ShardingManager{
		Shards: shards,
		Router: router,
	}

	// ============ Инициализация валидаторов ============
	peerAddresses := []string{
		"localhost:26656", // validator1
		"localhost:26657", // validator2
		"localhost:26658", // validator3
		"localhost:26659", // validator4
		"localhost:26660", // validator5
	}

	validators := []*pos.Validator{
		pos.NewValidatorWithAddress("validator1", peerAddresses[0], 2000),
		pos.NewValidatorWithAddress("validator2", peerAddresses[1], 1000),
		pos.NewValidatorWithAddress("validator3", peerAddresses[2], 1500),
		pos.NewValidatorWithAddress("validator4", peerAddresses[3], 1200),
		pos.NewValidatorWithAddress("validator5", peerAddresses[4], 800),
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
		"validator2": 1000,
		"validator3": 1500,
		"validator4": 1200,
		"validator5": 800,
	}

	guard := fiftyone.NewFiftyOnePercentGuard(validatorsMap)
	go guard.Monitor(30 * time.Second)

	// ============ Инициализация защиты от Sybil ============
	sybilGuard := sybil.NewSybilGuard([]string{
		"validator1",
		"validator2",
		"validator3",
		"validator4",
		"validator5",
	})

	peer.SetSybilGuard(sybilGuard)

	// ========== Инициализация аудита безопасности ==========
	auditor = audit.NewSecurityAuditor()

	// ============ Запуск REST API ============
	apiServer := api.NewAPIServer(chain, txPool, auditor)
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

	// ============ Инициализация говернанса ============
	// Создаем менеджер говернанса
	governanceManager := governance.NewGovernanceManager()

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
	switcher := manager.NewConsensusSwitcher(manager.ConsensusBFT)
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
    // Инициализация платежного канала
    channelStore := offchain.NewChannelStore()
    offchainAPI := &api.OffChainHandler{ChannelStore: channelStore}

    // Регистрация off-chain роутов
    apiServer.RegisterOffChainRoutes(offchainAPI)

    // Запуск тестового сценария
    runOffChainTestScenario(offchainAPI)
	// ============ Бесконечный цикл для поддержания работы ============
	select {}
}