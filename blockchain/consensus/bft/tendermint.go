package bft

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"blockchain/consensus/pos"
	"blockchain/crypto/signature"
	"blockchain/governance/reputation"
	"blockchain/network/gossip"
	"blockchain/network/p2p"
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
		log.Fatal("chain is nil")
	}
	if txPool == nil {
		log.Fatal("txPool is nil")
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

func (n *BFTNode) IsValidator(addr string) bool {
	for _, v := range n.ValidatorPool {
		if v.Address == addr {
			return true
		}
	}
	return false
}

// RunConsensusRound реализует полный раунд Tendermint-подобного консенсуса
func (n *BFTNode) RunConsensusRound() {
	// Синхронизируем высоту с другими узлами, если отстаём
	n.SyncHeight()

	// Инициализируем раунд, если его ещё нет или высота изменилась
	if n.CurrentRound == nil || n.CurrentRound.Height != n.Height {
		proposer := n.ValidatorPool.Select(n.Round)
		if proposer == nil {
			fmt.Println("❌ No proposer selected")
			return
		}

		// Проверяем, что пропосер действительно валидатор
		if !n.IsValidator(proposer.Address) {
			fmt.Printf("❌ Selected proposer is not a validator: %s\n", proposer.Address)
			n.Round++
			n.CurrentRound = nil
			return
		}

		repModule := reputation.NewReputationSystem()
		repScore := repModule.CalculateScore(proposer.Address, true)
		if repScore < 50 {
			fmt.Println("⚠️ Validator has low reputation, skipping")
			n.Round++
			n.CurrentRound = nil
			return
		}

		n.CurrentRound = NewRound(n.Height, n.Round, proposer.Address)
		fmt.Printf("🚀 Starting round %d for height %d. Proposer: %s\n", n.Round, n.Height, proposer.Address)
	}

	// Если мы — пропосер, создаём и отправляем блок
	if n.CurrentRound.Proposer == n.Address {
		if err := n.proposeBlock(n.CurrentRound); err != nil {
			fmt.Printf("❌ Failed to propose block: %v\n", err)
			n.Round++
			n.CurrentRound = nil
			return
		}
	} else {
		fmt.Printf("📬 Node is not proposer, waiting for proposal from %s\n", n.CurrentRound.Proposer)
	}

	// Ждём, пока не соберём голоса или не наступит таймаут
	time.Sleep(1 * time.Second)

	// Проверяем, не переключились ли мы на новый раунд
	if n.Round != n.CurrentRound.Round || n.Height != n.CurrentRound.Height {
		fmt.Println("❌ Attempt to sign old round")
		return
	}

	// 1. Prevote
	if err := n.signAndBroadcast(n.CurrentRound, gossip.StatePrevote); err != nil {
		fmt.Printf("❌ Failed to sign prevote: %v\n", err)
		n.Round++
		n.CurrentRound = nil
		return
	}

	time.Sleep(2 * time.Second)

	// 2. Precommit
	if err := n.signAndBroadcast(n.CurrentRound, gossip.StatePrecommit); err != nil {
		fmt.Printf("❌ Failed to sign precommit: %v\n", err)
		n.Round++
		n.CurrentRound = nil
		return
	}

	time.Sleep(1 * time.Second)

	// 3. Commit
	fmt.Printf("🗳 Total precommits received: %d\n", len(n.CurrentRound.Precommits))
	fmt.Printf("👥 Total validators: %d\n", len(n.ValidatorPool))

	if HasQuorum(n.CurrentRound.Precommits, n.ValidatorPool, n.CurrentRound.Round, n.CurrentRound.Height, n.CurrentRound.BlockHash) {
		if n.CurrentRound.ProposedBlock != nil {
			if err := n.processCommittedBlock(n.CurrentRound.ProposedBlock); err != nil {
				fmt.Printf("❌ Failed to process committed block: %v\n", err)
				n.Round++
				n.CurrentRound = nil
				return
			}
		} else {
			fmt.Println("❌ ProposedBlock is nil — cannot commit")
		}
	} else {
		fmt.Println("❌ Not enough precommits to commit")
		n.Round++
		n.CurrentRound = nil
		return
	}

	// Увеличиваем высоту только после успешного коммита
	n.Height++
	n.Round = 0
	n.CurrentRound = nil
}

func HasQuorum(votes map[string][]byte, validators []*pos.Validator, round, height int64, blockHash []byte) bool {
	totalPower := 0.0
	for _, v := range validators {
		totalPower += float64(v.Balance)
	}

	validVotes := 0.0
	data := []byte(fmt.Sprintf("prevote:%d:%d:%x", height, round, blockHash))
	hash := sha256.Sum256(data)

	for from, sig := range votes {
		for _, v := range validators {
			if v.Address == from {
				pubKey, err := signature.GetPublicKey(v.Address)
				if err != nil {
					continue
				}

				if ecdsa.VerifyASN1(pubKey, hash[:], sig) {
					validVotes += float64(v.Balance)
				}
				break
			}
		}
	}

	return validVotes > (2.0/3.0)*totalPower
}

func (n *BFTNode) SyncHeight() {
	maxHeight := int64(0)
	for _, addr := range n.Peers {
		peerHeight := n.GetPeerHeight(addr) // Реализуйте GetPeerHeight отдельно
		if peerHeight > maxHeight {
			maxHeight = peerHeight
		}
	}

	if maxHeight > n.Height {
		fmt.Printf("🔄 Node is behind. Syncing from height %d to %d\n", n.Height, maxHeight)
		n.Height = maxHeight
		n.Round = 0
		n.CurrentRound = nil
	}
}

func (n *BFTNode) GetPeerHeight(peerAddr string) int64 {
	// Устанавливаем TLS-соединение
	conn, err := tls.Dial("tcp", peerAddr, p2p.GenerateClientTLSConfig())
	if err != nil {
		fmt.Printf("❌ Failed to connect to peer %s: %v\n", peerAddr, err)
		return -1
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("❌ Failed to close connection to %s: %v\n", peerAddr, err)
		}
	}()

	// Создаём сообщение типа "status"
	msg := &gossip.ConsensusMessage{
		Type: gossip.MsgStatus,
		From: n.Address,
	}

	// Сериализуем сообщение
	data, err := msg.Encode()
	if err != nil {
		fmt.Printf("❌ Failed to encode status message: %v\n", err)
		return -1
	}

	// Отправляем запрос
	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("❌ Failed to send status request to %s: %v\n", peerAddr, err)
		return -1
	}

	// Ждём ответ
	buf := make([]byte, 4096)
	nBytes, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("❌ Failed to read response from %s: %v\n", peerAddr, err)
		return -1
	}

	// Десериализуем ответ
	resp, err := gossip.DecodeConsensusMessage(buf[:nBytes])
	if err != nil {
		fmt.Printf("❌ Failed to decode consensus message: %v\n", err)
		return -1
	}

	// Проверяем, что это сообщение с типом "status"
	if resp.Type != gossip.MsgStatus {
		fmt.Printf("❌ Unexpected message type: %s\n", resp.Type)
		return -1
	}

	// Возвращаем высоту
	return resp.Height
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
	round.BlockHash = []byte(block.CalculateHash())
	round.Step = gossip.StatePropose
	n.BroadcastSignedMessage(gossip.StatePropose, block.Serialize(), block.Signature)
	fmt.Printf("✅ Proposed block %s with %d transactions\n", block.Hash, len(validTxs))
	return nil
}

type ConsensusData struct {
	Type   gossip.MessageType
	Height int64
	Round  int64
}

func (n *BFTNode) signAndBroadcast(round *Round, msgType gossip.MessageType) error {
	data := &ConsensusData{
		Type:   msgType,
		Height: n.Height,
		Round:  round.Round,
	}

	// Сериализуем данные для подписи
	rawData := []byte(fmt.Sprintf("%s:%d:%d", data.Type, data.Height, data.Round))
	hash := sha256.Sum256(rawData) // хэшируем

	// Подписываем данные (возвращает DER)
	sig, err := n.Signer.Sign(hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign data: %v", err)
	}

	// Сохраняем подпись
	switch msgType {
	case gossip.StatePrevote:
		round.Prevotes[n.Address] = sig
	case gossip.StatePrecommit:
		round.Precommits[n.Address] = sig
	}

	// Отправляем данные и подпись
	n.BroadcastSignedMessage(msgType, hash[:], sig)
	fmt.Printf("🗳 %s from %s\n", msgType, n.Address)
	return nil
}

func (n *BFTNode) processCommittedBlock(blockData []byte) error {
	block := &blockchain.Block{}
	if err := block.Deserialize(blockData); err != nil {
		return fmt.Errorf("failed to deserialize block: %w", err)
	}

	// Проверяем, не был ли блок уже добавлен
	if n.Chain.HasBlock(block.Hash) {
		fmt.Printf("❌ Block %s already exists in chain\n", block.Hash)
		return nil
	}

	pubKey, err := signature.GetPublicKey(block.Validator)
	if err != nil {
		return fmt.Errorf("validator %s has no public key: %w", block.Validator, err)
	}

	if !signature.Verify(pubKey, block.SerializeWithoutSignature(), block.Signature) {
		return fmt.Errorf("invalid block signature")
	}

	// Проверяем каждую транзакцию на дублирование
	for _, tx := range block.Transactions {
		if n.Chain.HasTransaction(tx.ID) {
			fmt.Printf("❌ Transaction %s already exists in chain\n", tx.ID)
			continue
		}

		if !tx.Verify() {
			fmt.Printf("❌ Transaction %s is invalid\n", tx.ID)
			continue
		}
	}

	n.Chain.AddBlock(block)
	fmt.Printf("✅ Block added to chain: %s\n", block.Hash)

	for _, tx := range block.Transactions {
		n.TxPool.RemoveTransaction(tx.ID)
		fmt.Printf("🗑️ Removed transaction: %s\n", tx.ID)
	}

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
