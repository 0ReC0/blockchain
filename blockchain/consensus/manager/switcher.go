// consensus/manager/switcher.go
package manager

import (
	"blockchain/consensus/bft"
	"blockchain/consensus/pos"
	"blockchain/crypto/signature"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
	"fmt"
	"time"
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

func (cs *ConsensusSwitcher) startPoS(
	txPool *txpool.TransactionPool,
	chain *blockchain.Blockchain,
	validators []*pos.Validator,
	validatorPool pos.ValidatorPool,
	signer signature.Signer,
	peerAddresses []string,
) {
	for {
		selectedValidator := validatorPool.Select(int64(0))
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
	var newTransactions []*txpool.Transaction
	for _, tx := range transactions {
		if !chain.HasTransaction(tx.ID) {
			newTransactions = append(newTransactions, tx)
		}
	}
	if len(newTransactions) == 0 {
		return
	}
	prevBlock := chain.GetLatestBlock()
	if prevBlock == nil {
		fmt.Println("❌ Cannot create new block: no previous block found")
		return
	}
	block := &blockchain.Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevBlock.Hash,
		Transactions: newTransactions,
		Validator:    validator.Address,
	}
	block.Hash = block.CalculateHash()
	signatureBytes, _ := signer.Sign(block.SerializeWithoutSignature())
	block.Signature = signatureBytes
	chain.AddBlock(block)
	for _, tx := range newTransactions {
		txPool.RemoveTransaction(tx.ID)
	}
	fmt.Printf("✅ Block %d created by PoS validator %s\n", block.Index, validator.Address)
}

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
			fmt.Printf("✅ [%s] BFT Node started\n", validators[i].Address)
			bftNode.Start()
		}()
	}
}
