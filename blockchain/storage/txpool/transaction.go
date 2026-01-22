package txpool

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"blockchain/crypto/signature"
	"blockchain/governance/kyc"
)

// Global KYC manager instance
var kycManager *kyc.KYCManager

// SetKYCManager sets the global KYC manager instance
func SetKYCManager(manager *kyc.KYCManager) {
	kycManager = manager
}

type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    float64
	Fee       float64 // Добавлено: комиссия за транзакцию
	Timestamp int64
	Signature string
	IsPrivate bool
	Encrypted []byte
	PublicKey []byte
}

// Глобальная переменная для отслеживания размера пула транзакций
var transactionPoolSize int64

func NewTransaction(from, to string, amount float64) *Transaction {
	// Динамическое вычисление комиссии на основе размера пула транзакций
	baseFee := 0.001
	dynamicFee := baseFee

	// Увеличиваем комиссию при высокой нагрузке
	if transactionPoolSize > 1000 {
		dynamicFee = baseFee * 2
	} else if transactionPoolSize > 5000 {
		dynamicFee = baseFee * 5
	} else if transactionPoolSize > 10000 {
		dynamicFee = baseFee * 10
	}

	transactionPoolSize++

	return &Transaction{
		ID:        GenerateID(),
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       dynamicFee, // Добавляем динамическую комиссию
		Timestamp: time.Now().Unix(),
	}
}

func (t *Transaction) Serialize() []byte {
	temp := struct {
		From      string
		To        string
		Amount    float64
		Timestamp int64
	}{
		From:      t.From,
		To:        t.To,
		Amount:    t.Amount,
		Timestamp: t.Timestamp,
	}
	data, _ := json.Marshal(temp)
	return data
}

// Verify verifies the transaction signature and KYC status
func (t *Transaction) Verify() bool {
	// 1. Проверка наличия публичного ключа
	pubKey, err := signature.GetPublicKey(t.From)
	if err != nil {
		fmt.Printf("❌ Public key not found for %s: %v\n", t.From, err)
		return false
	}

	// 2. Проверка подписи транзакции
	sigBytes, err := hex.DecodeString(t.Signature)
	if err != nil {
		fmt.Printf("❌ Failed to decode signature hex: %v\n", err)
		return false
	}
	if !signature.Verify(pubKey, t.Serialize(), sigBytes) {
		fmt.Printf("❌ Signature verification failed for transaction %s\n", t.ID)
		return false
	}

	// 3. Проверка KYC (если доступен менеджер KYC)
	if kycManager != nil {
		kycStatus, _ := kycManager.CheckKYC(t.From)
		if kycStatus != kyc.Verified {
			fmt.Printf("❌ Transaction rejected: sender not verified (KYC status: %v)\n", kycStatus)
			return false
		}

		// 4. Проверка AML
		if kycManager.CheckSanctions(t.From) {
			fmt.Printf("❌ Transaction rejected: sender in sanctions list\n")
			return false
		}
	}

	return true
}
