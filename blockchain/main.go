// main.go
package main

import (
	"fmt"
	"time"

	// Консенсус
	"blockchain/consensus/bft"
	"blockchain/consensus/manager"
	"blockchain/consensus/pos"
	"blockchain/network/peer"

	// Хранилище
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"

	// Криптография
	"blockchain/crypto/signature"

	// Безопасность
	"blockchain/security/double_spend"
	"blockchain/security/fiftyone"
	"blockchain/security/sybil"

	// Сеть

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
	pubKey, err := signature.ParsePublicKey(signer.PublicKey())
	if err != nil {
		panic("❌ Failed to parse public key: " + err.Error())
	}
	signature.RegisterPublicKey(validators[0].Address, pubKey)
	signature.RegisterPublicKey(validators[1].Address, pubKey)

	// ============ Инициализация BFT-нод ============
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

	// ============ Инициализация ConsensusSwitcher ============
	switcher := manager.NewConsensusSwitcher(manager.ConsensusBFT)

	// ============ Инициализация защиты от 51% атак ============
	validatorsMap := map[string]int64{
		"validator1": 2000,
		"validator2": 1000,
	}
	guard := fiftyone.NewFiftyOnePercentGuard(validatorsMap)
	go guard.Monitor(30 * time.Second) // запуск мониторинга

	// ============ Инициализация защиты от Sybil ============
	sybilGuard := sybil.NewSybilGuard([]string{"validator1", "validator2"})
	peer.SetSybilGuard(sybilGuard)

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

	// ============ Запуск защиты от двойной траты ============
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

	// ============ Запуск консенсуса ============
	go func() {
		switcher.Run()
	}()

	// ============ Запуск BFT-узлов ============
	go func() {
		time.Sleep(2 * time.Second)
		bftNode.Start()
	}()
	go func() {
		time.Sleep(3 * time.Second)
		bftNode2.Start()
	}()

	fmt.Println("✅ Node started. Waiting for connections...")

	// ============ Бесконечный цикл для поддержания работы ============
	select {}
}
