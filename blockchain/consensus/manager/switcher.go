// consensus/manager/switcher.go
package manager

import (
	"fmt"
	"time"

	"blockchain/consensus/bft"
	"blockchain/consensus/pos"
	"blockchain/crypto/signature"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

type ConsensusType string

const (
	ConsensusPoS ConsensusType = "PoS"
	ConsensusBFT ConsensusType = "BFT"
)

type ConsensusSwitcher struct {
	Type ConsensusType
}

func NewConsensusSwitcher(t ConsensusType) *ConsensusSwitcher {
	return &ConsensusSwitcher{Type: t}
}

func (cs *ConsensusSwitcher) StartConsensus(
	txPool *txpool.TransactionPool,
	chain *blockchain.Blockchain,
	validators []*pos.Validator,
	validatorPool pos.ValidatorPool,
	signer signature.Signer,
	peerAddresses []string,
) {
	switch cs.Type {
	case ConsensusPoS:
		fmt.Println("🚀 Starting PoS consensus...")
		cs.startPoS(txPool, chain, validators, validatorPool, signer, peerAddresses)
	case ConsensusBFT:
		fmt.Println("🚀 Starting BFT consensus...")
		cs.startBFT(txPool, chain, validators, validatorPool, signer, peerAddresses)
	default:
		fmt.Println("❌ Unknown consensus type")
	}
}

// ===== PoS =====

func (cs *ConsensusSwitcher) startPoS(
	txPool *txpool.TransactionPool,
	chain *blockchain.Blockchain,
	validators []*pos.Validator,
	validatorPool pos.ValidatorPool,
	signer signature.Signer,
	peerAddresses []string,
) {
	for {
		// Выбираем валидатора через пул
		selectedValidator := validatorPool.Select()
		if selectedValidator == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		go func(v *pos.Validator) {
			fmt.Printf("⛏️ PoS Validator %s started\n", v.Address)
			cs.simulatePoSBlockCreation(chain, txPool, v, signer)
		}(selectedValidator)

		time.Sleep(10 * time.Second)
	}
}

func (cs *ConsensusSwitcher) simulatePoSBlockCreation(
	chain *blockchain.Blockchain,
	txPool *txpool.TransactionPool,
	validator *pos.Validator,
	signer signature.Signer,
) {
	transactions := txPool.GetTransactions(100)
	if len(transactions) == 0 {
		return
	}

	// Получаем последний блок из БД
	prevBlock := chain.GetLatestBlock()
	if prevBlock == nil {
		fmt.Println("❌ Cannot create new block: no previous block found")
		return
	}

	block := &blockchain.Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevBlock.Hash,
		Transactions: transactions,
		Validator:    validator.Address,
	}
	block.Hash = block.CalculateHash()
	signatureBytes, _ := signer.Sign(block.SerializeWithoutSignature())
	block.Signature = signatureBytes

	// Добавляем блок в БД
	chain.AddBlock(block)

	// Удаляем транзакции из пула
	for _, tx := range transactions {
		txPool.RemoveTransaction(tx.ID)
	}

	fmt.Printf("✅ Block %d created by PoS validator %s\n", block.Index, validator.Address)
}

// ===== BFT =====

func (cs *ConsensusSwitcher) startBFT(
	txPool *txpool.TransactionPool,
	chain *blockchain.Blockchain,
	validators []*pos.Validator,
	validatorPool pos.ValidatorPool,
	signer signature.Signer,
	peerAddresses []string,
) {
	for i := range validators {
		i := i
		go func() {
			bftNode := bft.NewBFTNode(
				validators[i].ID,
				validators[i],
				validatorPool,
				txPool,
				chain,
				signer,
				peerAddresses[i],
				peerAddresses,
			)
			bftNode.Start()
			fmt.Printf("✅ BFT Node %s started\n", validators[i].Address)
		}()
	}
}
