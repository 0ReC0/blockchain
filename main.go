// main.go
package main

import (
	"fmt"
	"time"

	// Уровень консенсуса
	"blockchain/consensus/bft"
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

	// ============ Инициализация валидаторов ============
	// Список всех пиров
	peerAddresses := []string{
		"localhost:26656", // validator1
		"localhost:26657", // validator2
	}

	validators := []*pos.Validator{
		pos.NewValidatorWithAddress("validator1", peerAddresses[0], 2000),
		pos.NewValidatorWithAddress("validator2", peerAddresses[1], 1000),
	}
	validatorPool := pos.NewValidatorPool(validators)

	// ============ Инициализация signer'а ============
	signer, err := signature.NewECDSASigner()
	if err != nil {
		panic("❌ Failed to create signer: " + err.Error())
	}
	// ============ Инициализация тестовой транзакции ============
	// Регистрация публичного ключа
	// 2. Получаем публичный ключ в виде []byte
	pubKeyBytes := signer.PublicKey()

	// 3. Десериализуем его в *ecdsa.PublicKey
	pubKey, err := signature.ParsePublicKey(pubKeyBytes)
	if err != nil {
		panic("Failed to parse public key: " + err.Error())
	}

	// 4. Регистрируем публичный ключ
	signature.RegisterPublicKey("A", pubKey)

	// 5. Создаём и подписываем транзакцию
	tx1 := txpool.NewTransaction("A", "B", 10)
	txBytes := tx1.Serialize()
	signatureBytes, err := signer.Sign(txBytes)
	if err != nil {
		panic("Failed to sign transaction: " + err.Error())
	}
	tx1.Signature = string(signatureBytes)

	// 6. Добавляем в пул
	txPool.AddTransaction(tx1)

	// ============ Инициализация BFT-ноды ============
	// Создаём BFT-ноду с адресом и пеерами
	bftNode := bft.NewBFTNode(
		"validator1",
		validators[0],
		*validatorPool,
		txPool,
		chain,
		signer,
		peerAddresses[0],
		peerAddresses,
	)
	bftNode2 := bft.NewBFTNode(
		"validator2",
		validators[1],
		*validatorPool,
		txPool,
		chain,
		signer,
		peerAddresses[1],
		peerAddresses,
	)
	// Регистрируем публичные ключи валидаторов
	signature.RegisterPublicKey(validators[0].Address, pubKey)
	signature.RegisterPublicKey(validators[1].Address, pubKey)

	// ============ Инициализация ConsensusSwitcher ============
	// switcher := manager.NewConsensusSwitcher(manager.ConsensusBFT)

	// ============ Запуск консенсуса через ConsensusSwitcher ============
	// go func() {
	// 	ticker := time.NewTicker(10 * time.Second)
	// 	for {
	// 		<-ticker.C
	// 		switcher.StartConsensus()
	// 	}
	// }()

	// ============ Запуск P2P сети ============
	go bft.StartTCPServer(bftNode)
	go bft.StartTCPServer(bftNode2)
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

	// Запуск первой ноды
	go func() {
		time.Sleep(2 * time.Second)
		bftNode.Start()
	}()

	// Запуск второй ноды
	go func() {
		time.Sleep(3 * time.Second)
		bftNode2.Start()
	}()
	fmt.Println("✅ Blockchain system started. Waiting for connections...")

	// ============ Бесконечный цикл для поддержания работы сервера ============
	select {}
}
