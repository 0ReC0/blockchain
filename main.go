package main

import (
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

	select {}
}
