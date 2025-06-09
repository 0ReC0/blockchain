package rollup

import (
	"../../../consensus/bft"
	"../../../storage/blockchain"
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
	// Отправляем батч транзакций
	if err := r.BFT.BroadcastMessage("batch", []byte(transactions)); err != nil {
		return err
	}
	return nil
}
