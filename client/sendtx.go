package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"
)

// =================== Транзакция ===================

type Transaction struct {
	ID        string  `json:"ID"`
	From      string  `json:"From"`
	To        string  `json:"To"`
	Amount    float64 `json:"Amount"`
	Timestamp int64   `json:"Timestamp"`
	Signature string  `json:"Signature"`
	IsPrivate bool    `json:"IsPrivate"`
}

func (tx *Transaction) String() string {
	return fmt.Sprintf("%s%s%s%f%d", tx.ID, tx.From, tx.To, tx.Amount, tx.Timestamp)
}

// =================== Генерация ключей ===================

func GenerateKeys() (string, string, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	pubKey := &privKey.PublicKey
	pubKeyBytes := elliptic.MarshalCompressed(pubKey, pubKey.X, pubKey.Y)

	return hex.EncodeToString(privKey.D.Bytes()), hex.EncodeToString(pubKeyBytes), nil
}

// =================== Подпись ===================

func CalculateTxHash(tx *Transaction) []byte {
	hash := sha256.Sum256([]byte(tx.String()))
	return hash[:]
}

func SignTransaction(tx *Transaction, privKeyHex string) (string, error) {
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return "", err
	}

	privKey := new(ecdsa.PrivateKey)
	privKey.Curve = elliptic.P256()
	privKey.D = new(big.Int).SetBytes(privKeyBytes)
	privKey.PublicKey.X, privKey.PublicKey.Y = elliptic.P256().ScalarBaseMult(privKeyBytes)

	hash := CalculateTxHash(tx)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return "", err
	}

	sig := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(sig), nil
}

// =================== Проверка подписи ===================

func VerifyTransaction(tx *Transaction, pubKeyHex string) bool {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false
	}

	curve := elliptic.P256()
	x, y := elliptic.UnmarshalCompressed(curve, pubKeyBytes)
	if x == nil {
		return false
	}

	pubKey := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	hash := CalculateTxHash(tx)
	sigBytes, err := hex.DecodeString(tx.Signature)
	if err != nil || len(sigBytes) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	return ecdsa.Verify(pubKey, hash, r, s)
}

// =================== Отправка на API ===================

func SendTransaction(tx *Transaction) error {
	url := "http://localhost:8081/transactions"

	body, _ := json.Marshal(tx)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("📡 Response: %d\n%s\n", resp.StatusCode, string(bodyBytes))
	return nil
}

// =================== Main ===================

func main() {
	// 1. Генерируем ключи
	privKey, pubKey, err := GenerateKeys()
	if err != nil {
		panic(err)
	}

	fmt.Printf("🔐 Private Key: %s\n", privKey)
	fmt.Printf("📘 Public Key:  %s\n", pubKey)

	// 2. Создаем транзакцию
	tx := &Transaction{
		ID:        "tx_001",
		From:      "validator1",
		To:        "validator2",
		Amount:    50.0,
		Timestamp: time.Now().Unix(),
		IsPrivate: false,
	}

	// 3. Подписываем
	sig, err := SignTransaction(tx, privKey)
	if err != nil {
		panic(err)
	}
	tx.Signature = sig

	// 4. Проверяем
	if !VerifyTransaction(tx, pubKey) {
		panic("❌ Signature verification failed")
	}

	// 5. Выводим JSON
	jsonTx, _ := json.MarshalIndent(tx, "", "  ")
	fmt.Printf("\n📤 Transaction JSON:\n%s\n", string(jsonTx))

	// 6. Отправляем
	err = SendTransaction(tx)
	if err != nil {
		panic(err)
	}
}
