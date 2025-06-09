package bft

// реализация BFT (упрощённый Tendermint)

import (
	"fmt"
	"time"

	"../../network/gossip"
	"../../network/peer"
	"../../storage/blockchain"
	"../../storage/txpool"
	"../pos"
)

type BFTNode struct {
	Address       string
	Validator     *pos.Validator
	ValidatorPool pos.ValidatorPool
	Height        int64
	Round         int64
	TxPool        *txpool.TransactionPool
	Chain         *blockchain.Blockchain
}

func NewBFTNode(
	addr string,
	val *pos.Validator,
	validatorPool pos.ValidatorPool,
	txPool *txpool.TransactionPool,
	chain *blockchain.Blockchain,
) *BFTNode {
	return &BFTNode{
		Address:       addr,
		Validator:     val,
		ValidatorPool: validatorPool,
		Height:        0,
		Round:         0,
		TxPool:        txPool,
		Chain:         chain,
	}
}

func (n *BFTNode) Start() {
	for {
		n.RunConsensusRound()
		n.Height++
	}
}

func (n *BFTNode) BroadcastMessage(msgType MessageType, data []byte) {
	msg := &gossip.ConsensusMessage{
		Type:   gossip.MessageType(msgType), // Преобразуем тип сообщения
		Height: n.Height,
		Round:  n.Round,
		From:   n.Address,
		Data:   data,
	}

	// Преобразуем ValidatorPool в []*peer.Peer
	var peers []*peer.Peer
	for _, validator := range n.ValidatorPool {
		peers = append(peers, &peer.Peer{
			ID:   validator.Address,
			Addr: "unknown", // можно заменить на реальный адрес, если он есть
		})
	}

	// Передаем преобразованный список пиров
	if err := gossip.BroadcastConsensusMessage(peers, msg); err != nil {
		fmt.Printf("Failed to broadcast message: %v\n", err)
	}
}

func (n *BFTNode) RunConsensusRound() {
	// Выбор пропосера
	proposer := n.ValidatorPool.Select()
	round := NewRound(n.Height, n.Round, proposer.Address)

	fmt.Printf("Starting round %d for height %d. Proposer: %s\n", n.Round, n.Height, proposer.Address)

	// 1. Propose
	if proposer.Address == n.Address {
		transactions := n.TxPool.GetTransactions(100)
		if len(transactions) == 0 {
			fmt.Println("No transactions to propose")
			return
		}

		prevBlock := n.Chain.Blocks[len(n.Chain.Blocks)-1]
		block := &blockchain.Block{
			Index:        prevBlock.Index + 1,
			Timestamp:    time.Now().Unix(),
			PrevHash:     prevBlock.Hash,
			Transactions: transactions,
			Validator:    n.Address,
		}
		block.Hash = block.CalculateHash()

		round.ProposedBlock = block
		round.Step = MsgPropose

		// Отправляем предложение
		n.BroadcastMessage(MsgPropose, block.Serialize())
		fmt.Printf("Proposed block %s with %d transactions\n", block.Hash, len(transactions))
	}

	time.Sleep(1 * time.Second)

	// 2. Prevote
	round.Prevotes[n.Address] = []byte("prevote")
	n.BroadcastMessage(MsgPrevote, []byte("prevote"))
	fmt.Printf("Prevote from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 3. Precommit
	round.Precommits[n.Address] = []byte("precommit")
	n.BroadcastMessage(MsgPrecommit, []byte("precommit"))
	fmt.Printf("Precommit from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 4. Commit
	if len(round.Precommits) >= 2 {
		if round.ProposedBlock != nil {
			n.Chain.Blocks = append(n.Chain.Blocks, round.ProposedBlock)

			// Очищаем транзакции из пула
			for _, tx := range round.ProposedBlock.Transactions {
				n.TxPool.RemoveTransaction(tx.ID)
			}

			n.BroadcastMessage(MsgCommit, round.ProposedBlock.Serialize())
			fmt.Printf("Block committed: %s\n", round.ProposedBlock.Hash)
		}
	}
}
