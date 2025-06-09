// main.go
package main

import (
	"fmt"
	"time"

	// Основные модули
	"./consensus/bft"
	"./consensus/manager"
	"./consensus/pos"
	"./network/p2p"

	// Уровень хранения
	"./storage/blockchain"
	"./storage/sharding"
	"./storage/txpool"

	// Модули безопасности
	"./security/double_spend"

	// Модули масштабируемости
	"./scalability/parallel"

	// Смарт-контракты

	// Интеграция
	"./integration/api"
	"./integration/bank"
	"./integration/crosschain"

	// Говернанс
	"./governance/reputation"
	"./governance/upgrade"
	"./governance/voting"
)

func main() {
	// ============ Инициализация блокчейна и пула транзакций ============
	chain := blockchain.NewBlockchain()
	txPool := txpool.NewTransactionPool()
	tx1 := txpool.NewTransaction("A", "B", 10)
	txPool.AddTransaction(tx1)

	// ============ Инициализация сетевого уровня ============
	go p2p.StartNetwork()

	// ============ Инициализация API ============
	apiServer := api.NewAPIServer(chain, txPool)
	go apiServer.Start(":8081")

	// ============ Инициализация консенсуса ============
	posManager := manager.NewConsensusManager(manager.ConsensusPoS)
	bftManager := manager.NewConsensusManager(manager.ConsensusBFT)

	// ============ Инициализация BFT-ноды ============
	val := pos.NewValidator("validator1", 2000)
	bftNode := bft.NewBFTNode(
		"validator1",
		val,
		pos.NewValidatorPool([]*pos.Validator{val}),
		txPool,
		chain,
	)

	// ============ Инициализация говернанса ============
	upgradeMgr := upgrade.NewUpgradeManager("v1.0.0")
	upgradeMgr.ProposeUpgrade("v2.0.0", "Switch to BFT", time.Now().Add(24*time.Hour))
	upgradeMgr.ApproveUpgrade()
	_ = upgradeMgr.ApplyUpgrade()

	// ============ Инициализация шардов ============
	shardMgr := sharding.NewShardManager()
	shardMgr.CreateShard("0")

	// ============ Инициализация безопасности ============
	double_spend.InitSecurity()

	// ============ Инициализация масштабируемости ============
	executor := parallel.NewParallelExecutor(4)
	executor.ExecuteTransactions(txPool.GetTransactions(100), chain)

	// ============ Инициализация межблокчейновой интеграции ============
	chainA := blockchain.NewBlockchain()
	chainB := blockchain.NewBlockchain()
	bridge := crosschain.NewChainBridge(chainA, chainB)

	// ============ Инициализация банковского шлюза ============
	bank := bank.NewBankGateway("api-key", "https://bank-api.com")

	// ============ Инициализация голосования ============
	voting := voting.NewVotingModule()
	reputation := reputation.NewReputationModule()

	// ============ Запуск всех компонентов ============
	go posManager.Run()
	go bftManager.Run()
	go bftNode.Start()

	fmt.Println("Blockchain system started. Waiting for connections...")

	// ============ Бесконечный цикл для поддержания работы сервера ============
	select {}
}
