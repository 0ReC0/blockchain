package bft

import (
	"fmt"
	"time"

	"blockchain/consensus/pos"
	"blockchain/crypto/signature"
	"blockchain/governance/reputation"
	"blockchain/network/gossip"
	"blockchain/network/peer"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

// BFTNode — узел, участвующий в консенсусе Tendermint
type BFTNode struct {
	ID            string
	Address       string
	Validator     *pos.Validator
	ValidatorPool pos.ValidatorPool
	Peers         []string
	Height        int64
	Round         int64
	State         gossip.MessageType
	TxPool        *txpool.TransactionPool
	Chain         *blockchain.Blockchain
	Signer        signature.Signer
	CurrentRound  *Round
}

// NewBFTNode создаёт новый экземпляр BFTNode
func NewBFTNode(
	id string,
	validator *pos.Validator,
	validatorPool pos.ValidatorPool,
	txPool *txpool.TransactionPool,
	chain *blockchain.Blockchain,
	signer signature.Signer,
	address string,
	peers []string,
) *BFTNode {
	if chain == nil {
		panic("chain is nil")
	}
	if txPool == nil {
		panic("txPool is nil")
	}

	return &BFTNode{
		ID:            id,
		Address:       address,
		Validator:     validator,
		ValidatorPool: validatorPool,
		Peers:         peers,
		Height:        0,
		Round:         0,
		State:         gossip.StatePropose,
		TxPool:        txPool,
		Chain:         chain,
		Signer:        signer,
		CurrentRound:  NewRound(0, 0, ""),
	}
}

// Start запускает цикл консенсуса
func (n *BFTNode) Start() {
	go StartTCPServer(n)

	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		n.RunConsensusRound()
		n.Height++
	}
}

// RunConsensusRound реализует полный раунд Tendermint-подобного консенсуса
func (n *BFTNode) RunConsensusRound() {
	// 1. Выбор пропосера
	proposer := n.ValidatorPool.Select()
	if proposer == nil {
		fmt.Println("❌ No proposer selected")
		return
	}

	repModule := reputation.NewReputationSystem()

	// Обновляем репутацию перед выбором
	for _, v := range n.ValidatorPool {
		repModule.UpdateReputation(v.Address, 1.0)
	}

	repScore := repModule.CalculateScore(proposer.Address, true)
	if repScore < 50 {
		fmt.Println("⚠️ Validator has low reputation, skipping")
		return
	}

	// Инициализируем раунд
	round := NewRound(n.Height, n.Round, proposer.Address)
	n.CurrentRound = round

	fmt.Printf("🚀 Starting round %d for height %d. Proposer: %s\n", n.Round, n.Height, proposer.Address)

	// 2. Propose (только если мы — пропосер)
	if proposer.Address == n.Address {
		if err := n.proposeBlock(round); err != nil {
			fmt.Printf("❌ Failed to propose block: %v\n", err)
			repModule.UpdateReputation(n.Address, -10) // Снижаем репутацию
			return
		}
		repModule.UpdateReputation(n.Address, 10) // Повышаем за успешное предложение
	} else {
		fmt.Printf("📬 Node is not proposer, waiting for proposal from %s\n", proposer.Address)
	}

	time.Sleep(1 * time.Second)

	// 3. Prevote
	if err := n.signAndBroadcast(round, gossip.StatePrevote); err != nil {
		fmt.Printf("❌ Failed to sign prevote: %v\n", err)
		repModule.UpdateReputation(n.Address, -5)
		return
	}

	time.Sleep(3 * time.Second)

	// 4. Precommit
	if err := n.signAndBroadcast(round, gossip.StatePrecommit); err != nil {
		fmt.Printf("❌ Failed to sign precommit: %v\n", err)
		repModule.UpdateReputation(n.Address, -5)
		return
	}

	time.Sleep(1 * time.Second)

	// 5. Commit
	if HasQuorum(round.Precommits, len(n.ValidatorPool)) {
		if round.ProposedBlock != nil {
			if err := n.processCommittedBlock(round.ProposedBlock); err != nil {
				fmt.Printf("❌ Failed to process committed block: %v\n", err)
				repModule.UpdateReputation(n.Address, -10)
				return
			}
			repModule.UpdateReputation(n.Address, 10) // Повышение за успешный коммит
		} else {
			fmt.Println("❌ ProposedBlock is nil — cannot commit")
		}
	} else {
		fmt.Println("❌ Not enough precommits to commit")
	}
}

func (n *BFTNode) proposeBlock(round *Round) error {
	if n.Chain == nil {
		return fmt.Errorf("chain is nil")
	}

	if n.Chain.DB() == nil {
		return fmt.Errorf("chain.db is nil")
	}

	transactions := n.TxPool.GetTransactions(100)
	if len(transactions) == 0 {
		return fmt.Errorf("no transactions to propose")
	}

	var validTxs []*txpool.Transaction
	for _, tx := range transactions {
		if tx.Verify() {
			validTxs = append(validTxs, tx)
		}
	}

	if len(validTxs) == 0 {
		return fmt.Errorf("no valid transactions to propose")
	}

	// Получаем последний блок
	prevBlock := n.Chain.GetLatestBlock()
	if prevBlock == nil {
		return fmt.Errorf("chain is empty or invalid")
	}

	block := blockchain.NewBlock(
		prevBlock.Index+1,
		prevBlock.Hash,
		validTxs,
		n.Address,
	)

	signatureBytes, err := n.Signer.Sign(block.SerializeWithoutSignature())
	if err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}
	block.Signature = signatureBytes

	round.ProposedBlock = block.Serialize()
	round.Step = gossip.StatePropose
	n.BroadcastSignedMessage(gossip.StatePropose, block.Serialize(), block.Signature)
	fmt.Printf("✅ Proposed block %s with %d transactions\n", block.Hash, len(validTxs))
	return nil
}

func (n *BFTNode) signAndBroadcast(round *Round, msgType gossip.MessageType) error {
	data := []byte(fmt.Sprintf("%s:%d:%d", msgType, n.Height, n.Round))
	sig, err := n.Signer.Sign(data)
	if err != nil {
		return err
	}

	switch msgType {
	case gossip.StatePrevote:
		round.Prevotes[n.Address] = sig
	case gossip.StatePrecommit:
		round.Precommits[n.Address] = sig
	}

	n.BroadcastSignedMessage(msgType, data, sig)
	fmt.Printf("🗳 %s from %s\n", msgType, n.Address)

	return nil
}

func (n *BFTNode) processCommittedBlock(blockData []byte) error {
	block := &blockchain.Block{}
	if err := block.Deserialize(blockData); err != nil {
		return fmt.Errorf("failed to deserialize block: %w", err)
	}

	pubKey, err := signature.GetPublicKey(block.Validator)
	if err != nil {
		return fmt.Errorf("validator %s has no public key: %w", block.Validator, err)
	}

	if !signature.Verify(pubKey, block.SerializeWithoutSignature(), block.Signature) {
		return fmt.Errorf("invalid block signature")
	}

	n.Chain.AddBlock(block)
	fmt.Printf("✅ Block added to chain: %s\n", block.Hash)

	for _, tx := range block.Transactions {
		n.TxPool.RemoveTransaction(tx.ID)
		fmt.Printf("🗑️ Removed transaction: %s\n", tx.ID)
	}

	commitSig, err := n.Signer.Sign(block.SerializeWithoutSignature())
	if err != nil {
		return fmt.Errorf("failed to sign commit: %w", err)
	}

	n.BroadcastSignedMessage(gossip.StateCommit, block.SerializeWithoutSignature(), commitSig)
	fmt.Printf("✅ Block committed: %s\n", block.Hash)

	return nil
}

func (n *BFTNode) BroadcastSignedMessage(msgType gossip.MessageType, data, signature []byte) {
	// Конвертируем []string в []*peer.Peer
	peers := make([]*peer.Peer, len(n.Peers))
	for i, addr := range n.Peers {
		peers[i] = &peer.Peer{
			Addr: addr,
			ID:   addr,
		}
	}

	gossip.BroadcastSignedConsensusMessage(peers, &gossip.SignedConsensusMessage{
		Type:      msgType,
		Height:    n.Height,
		Round:     n.Round,
		From:      n.Address,
		Data:      data,
		Signature: signature,
	})
}
