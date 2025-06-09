package main

import (
	"fmt"

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
}
