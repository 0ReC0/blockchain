// main.go
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
	"blockchain/governance/upgrade"
)

func main() {
	fmt.Println("🚀 Starting Minimal Blockchain Node...")

	// ============ Инициализация хранилища ============
	chain := blockchain.NewBlockchain()

	txPool := txpool.NewTransactionPool()

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

		// Для демонстрации — покажем адрес и стейк
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
	go guard.Monitor(30 * time.Second) // запуск мониторинга

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
	auditor := audit.NewSecurityAuditor()

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
		Message:   "Blockchain node started successfully",
		NodeID:    "validator1",
		Severity:  "INFO",
	})

	// ========== Используем аудит в других компонентах ==========
	double_spend.SetAuditor(auditor) // Передаем аудит в защиту от двойной траты
	fiftyone.SetAuditor(auditor)     // Передаем аудит в защиту от 51% атак
	sybil.SetAuditor(auditor)        // Передаем аудит в защиту от Sybil

	// ============ Инициализация говернанса ============
	upgradeMgr := upgrade.NewUpgradeManager("v1.0.0")
	upgradeMgr.ProposeUpgrade("v2.0.0", "Switch to BFT", time.Now().Add(24*time.Hour))
	if err := upgradeMgr.ApproveUpgrade(); err != nil {
		fmt.Println("⚠️ Approval failed:", err)
	}
	if err := upgradeMgr.ApplyUpgrade(); err != nil {
		fmt.Println("⚠️ Upgrade failed:", err)
	}

	// ============ Запуск консенсуса ============
	switcher := manager.NewConsensusSwitcher(manager.ConsensusBFT)

	switcher.StartConsensus(
		txPool,
		chain,
		validators,
		*validatorPool,
		signer,
		peerAddresses,
	)

	fmt.Println("✅ Node started. Waiting for connections...")

	// ============ Бесконечный цикл для поддержания работы ============
	select {}
}
