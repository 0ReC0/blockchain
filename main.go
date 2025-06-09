package main

import (
	"fmt"
	"log"
	"time"

	bft "./consensus/bft"
	manager "./consensus/manager"
	pos "./consensus/pos"
	p2p "./network/p2p"
	rpc "./network/rpc"
)

func main() {
	go p2p.StartNetwork()
	rpc.StartRPCServer(":8080")

	// Запуск PoS-консенсуса
	posManager := manager.NewConsensusManager(manager.ConsensusPoS)
	go posManager.Run()

	// Запуск BFT-консенсуса
	bftManager := manager.NewConsensusManager(manager.ConsensusBFT)
	go bftManager.Run()

	// Запуск BFT-ноды
	val := pos.NewValidator("validator1", 2000)
	bftNode := bft.NewBFTNode("validator1", val)
	go bftNode.Start()

	// Создаем говернанс
	upgradeMgr := manager.NewUpgradeManager()
	upgradeMgr.SubmitUpgradeProposal("Update consensus protocol", "Switch to faster BFT", "validator1")

	// Голосование
	upgradeMgr.Governance.VoteOnProposal("upgrade-1", "validator1", "yes")
	upgradeMgr.Governance.VoteOnProposal("upgrade-1", "validator2", "yes")
	upgradeMgr.ApproveUpgrade("upgrade-1")

	// Создаем блокчейн
	chain := blockchain.NewBlockchain()

	// Добавляем транзакции
	txPool := txpool.NewTransactionPool()
	tx1 := txpool.NewTransaction("A", "B", 10)
	tx2 := txpool.NewTransaction("B", "C", 5)
	txPool.AddTransaction(tx1)
	txPool.AddTransaction(tx2)

	// Создаем шарды
	shardMgr := sharding.NewShardManager()
	shardMgr.CreateShard("0")
	shardMgr.CreateShard("1")
	shardMgr.CreateShard("2")

	// Добавляем блок
	validator := "validator1"
	transactions := txPool.GetTransactions(2)
	block := blockchain.NewBlock(1, chain.Blocks[0].Hash, transactions, validator)
	chain.AddBlock(transactions, validator)

	// Выводим цепочку
	for _, block := range chain.Blocks {
		fmt.Printf("Block %d: %s\n", block.Index, block.Hash)
	}

	select {}

	// Инициализация блокчейна
	chain := blockchain.NewBlockchain()

	// Инициализация пула транзакций
	txPool := txpool.NewTransactionPool()
	tx1 := txpool.NewTransaction("A", "B", 10)
	tx2 := txpool.NewTransaction("B", "C", 5)
	txPool.AddTransaction(tx1)
	txPool.AddTransaction(tx2)

	// Инициализация BFT-ноды
	val := pos.NewValidator("validator1", 2000)
	bftNode := bft.NewBFTNode("validator1", val)

	// Запуск консенсуса
	go func() {
		for {
			bftNode.RunConsensusRound(txPool, chain)
			bftNode.Height++
			time.Sleep(10 * time.Second)
		}
	}()

	// Выводим цепочку
	for {
		time.Sleep(30 * time.Second)
		for _, block := range chain.Blocks {
			fmt.Printf("Block %d: %s\n", block.Index, block.Hash)
		}
	}

	// Инициализация шифрования
	aesEncryptor := &encryption.AESEncryptor{}
	key := private_tx.GenerateKey("my-secret-password")

	// Создание приватной транзакции
	tx, _ := private_tx.NewPrivateTransaction("A", "B", 10, aesEncryptor, key)

	// Расшифровка
	decrypted, _ := tx.Decrypt(aesEncryptor, key)
	fmt.Println("Decrypted:", string(decrypted))

	// Пул приватных транзакций
	pool := shielded.NewShieldedPool()
	pool.AddTransaction(tx)

	// Блок с приватными транзакциями
	baseBlock := &blockchain.Block{
		Index:    1,
		PrevHash: "0",
		Hash:     "1",
	}
	shieldedBlock := shielded.NewShieldedBlock(baseBlock, pool.GetTransactions(1))
	fmt.Printf("Shielded block: %+v\n", shieldedBlock)

	// Инициализация шардов
	shardMgr := sharding.NewShardManager()
	shardMgr.CreateShard("0")
	shardMgr.CreateShard("1")
	shardMgr.CreateShard("2")

	// Инициализация пула транзакций
	txPool := txpool.NewTransactionPool()
	tx1 := txpool.NewTransaction("A", "B", 10)
	tx2 := txpool.NewTransaction("B", "C", 5)
	txPool.AddTransaction(tx1)
	txPool.AddTransaction(tx2)

	// Параллельная обработка
	executor := parallel.NewParallelExecutor(4)
	executor.ExecuteTransactions(txPool.GetTransactions(100), chain)

	// Шардинг
	shard := shardMgr.GetShardForAddress("A")
	for _, tx := range txPool.GetTransactions(100) {
		shard.AddTransaction(tx)
	}
	shard.FinalizeBlock()

	handler := execution.NewContractHandler()

	// Деплоим токен
	tokenAddr := handler.DeployERC20("MyToken", "MTK", 18, 1_000_000)
	fmt.Println("Token deployed at:", tokenAddr)

	// Вызываем метод
	result, err := handler.CallERC20(tokenAddr, "transfer", "0x123", "0x456", uint64(100))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transfer result:", result)

	// Инициализация блокчейна
	chainA := blockchain.NewBlockchain()
	chainB := blockchain.NewBlockchain()

	// Инициализация пула транзакций
	txPool := txpool.NewTransactionPool()

	// REST API
	apiServer := api.NewAPIServer(chainA, txPool)
	go apiServer.Start(":8080")

	// Cross-Chain
	bridge := crosschain.NewChainBridge(chainA, chainB)
	orcl := crosschain.NewCrossChainOracle()
	orcl.Bridges = append(orcl.Bridges, bridge)

	// Запускаем оракул
	go orcl.MonitorChains()

	// Банковский шлюз
	bank := bank.NewBankGateway("api-key", "https://bank-api.com")
	adapter := bank.NewBankAdapter(bank)

	// Депозит
	tx := txpool.NewTransaction("bank-reserve", "user1", 1000)
	bankTxID, _ := adapter.HandleDeposit(tx)
	fmt.Println("Deposit processed:", bankTxID)

	// Ожидаем
	select {}
}
