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
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		n.RunConsensusRound()
		n.Height++
	}
}

// RunConsensusRound реализует полный раунд Tendermint-подобного консенсуса
func (n *BFTNode) RunConsensusRound() {
	// Выбор пропосера
	proposer := n.ValidatorPool.Select()
	if proposer == nil {
		fmt.Println("No proposer selected")
		return
	}

	repModule := reputation.NewReputationSystem()

	// Обновляем репутацию перед выбором
	for _, v := range n.ValidatorPool {
		repModule.UpdateReputation(v.Address, 1.0)
	}

	repScore := repModule.CalculateScore(proposer.Address, true)
	if repScore < 50 {
		fmt.Println("Validator has low reputation, skipping")
		return
	}

	round := NewRound(n.Height, n.Round, proposer.Address)
	n.CurrentRound = round

	fmt.Printf("Starting round %d for height %d. Proposer: %s\n", n.Round, n.Height, proposer.Address)

	// 1. Propose
	if proposer.Address == n.Address {
		transactions := n.TxPool.GetTransactions(100)
		if len(transactions) == 0 {
			fmt.Println("No transactions to propose")
			return
		}

		// Проверяем подписи транзакций
		var validTxs []*txpool.Transaction
		for _, tx := range transactions {
			if tx.Verify() {
				validTxs = append(validTxs, tx)
			} else {
				fmt.Printf("Invalid transaction: %s\n", tx.ID)
			}
		}

		if len(validTxs) == 0 {
			fmt.Println("No valid transactions to propose")
			return
		}

		// Создаем блок
		prevBlock := n.Chain.Blocks[len(n.Chain.Blocks)-1]
		block := &blockchain.Block{
			Index:        prevBlock.Index + 1,
			Timestamp:    time.Now().Unix(),
			PrevHash:     prevBlock.Hash,
			Transactions: validTxs,
			Validator:    n.Address,
		}
		block.Hash = block.CalculateHash()

		// Подписываем блок
		signatureBytes, err := n.Signer.Sign(block.Serialize())
		if err != nil {
			fmt.Printf("Failed to sign block: %v\n", err)
			return
		}
		block.Signature = signatureBytes

		round.ProposedBlock = block.Serialize()
		round.Step = gossip.StatePropose

		// Отправляем предложение
		n.BroadcastSignedMessage(gossip.StatePropose, block.Serialize(), block.Signature)
		fmt.Printf("Proposed block %s with %d transactions\n", block.Hash, len(validTxs))
	}

	time.Sleep(1 * time.Second)

	// 2. Prevote
	// Подписываем prevote
	prevoteData := []byte(fmt.Sprintf("prevote:%d:%d", n.Height, n.Round))
	prevoteSig, err := n.Signer.Sign(prevoteData)
	if err != nil {
		fmt.Printf("Failed to sign prevote: %v\n", err)
		return
	}

	round.Prevotes[n.Address] = prevoteSig
	n.BroadcastSignedMessage(gossip.StatePrevote, prevoteData, prevoteSig)
	fmt.Printf("Prevote from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 3. Precommit
	// Подписываем precommit
	precommitData := []byte(fmt.Sprintf("precommit:%d:%d", n.Height, n.Round))
	precommitSig, err := n.Signer.Sign(precommitData)
	if err != nil {
		fmt.Printf("Failed to sign prevote: %v\n", err)
		return
	}

	round.Precommits[n.Address] = precommitSig
	n.BroadcastSignedMessage(gossip.StatePrecommit, precommitData, precommitSig)
	fmt.Printf("Precommit from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 4. Commit
	if len(round.Precommits) >= 2 {
		if round.ProposedBlock != nil {
			// Десериализуем блок
			block := &blockchain.Block{}
			if err := block.Deserialize(round.ProposedBlock); err != nil {
				fmt.Printf("Failed to deserialize block: %v\n", err)
				return
			}

			pubKey, err := signature.GetPublicKey(block.Validator)
			if err != nil {
				fmt.Printf("Validator %s has no public key: %v\n", block.Validator, err)
				return
			}
			// Проверяем подпись блока
			if !signature.Verify(pubKey, block.Serialize(), block.Signature) {
				fmt.Println("Invalid block signature")
				return
			}

			// Добавляем блок в цепочку
			n.Chain.Blocks = append(n.Chain.Blocks, block)

			// Очищаем транзакции из пула
			for _, tx := range block.Transactions {
				n.TxPool.RemoveTransaction(tx.ID)
			}

			// Подписываем коммит

			commitSig, err := n.Signer.Sign(block.Serialize())
			if err != nil {
				fmt.Printf("Failed to sign commit: %v\n", err)
				return
			}
			n.BroadcastSignedMessage(gossip.StateCommit, block.Serialize(), commitSig)
			fmt.Printf("Block committed: %s\n", block.Hash)
		}
	}
}

func (n *BFTNode) BroadcastSignedMessage(msgType gossip.MessageType, data, signature []byte) {
	// Конвертируем []string в []*peer.Peer
	peers := make([]*peer.Peer, len(n.Peers))
	for i, addr := range n.Peers {
		peers[i] = &peer.Peer{
			Addr: addr,
			ID:   addr, // или генерируй ID как-то иначе
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
