package rollup

import (
	"blockchain/consensus/bft"
	"blockchain/network/gossip"
	"blockchain/network/peer"
	"blockchain/storage/blockchain"
)

type OptimisticRollup struct {
	Chain *blockchain.Blockchain
	BFT   *bft.BFTNode
}

func NewOptimisticRollup(chain *blockchain.Blockchain, bftNode *bft.BFTNode) *OptimisticRollup {
	return &OptimisticRollup{
		Chain: chain,
		BFT:   bftNode,
	}
}

func (r *OptimisticRollup) SubmitBatch(transactions []string) error {
	// Преобразуем transactions в формат, понятный gossip
	msg := &gossip.GossipMessage{
		Type: gossip.MsgTx,
		From: r.BFT.Address,
		Data: []byte(transactions[0]), // упрощённый пример
	}

	// Преобразуем ValidatorPool в []*peer.Peer
	var peers []*peer.Peer
	for _, validator := range r.BFT.ValidatorPool {
		peers = append(peers, &peer.Peer{
			ID:   validator.Address,
			Addr: "unknown", // можно улучшить, получая из реестра пиров
		})
	}

	// Используем существующую функцию
	if err := gossip.Broadcast(peers, msg); err != nil {
		return err
	}
	return nil
}
