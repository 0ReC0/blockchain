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

	// 🔐 Ручная кодировка DER
	sig, err := MarshalECDSASignature(r, s)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sig), nil
}
func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Подготавливаем байты с учетом ASN.1 INTEGER
	// Если старший бит установлен, добавляем префикс 0x00
	rPrefix := 0
	if len(rBytes) > 0 && rBytes[0] >= 0x80 {
		rPrefix = 1
	}

	sPrefix := 0
	if len(sBytes) > 0 && sBytes[0] >= 0x80 {
		sPrefix = 1
	}

	// Вычисляем длину
	length := 6 + len(rBytes) + len(sBytes) + rPrefix + sPrefix

	// Создаем буфер
	sig := make([]byte, length)

	// ASN.1 SEQUENCE
	sig[0] = 0x30
	sig[1] = byte(length - 2) // Длина последовательности

	// r
	sig[2] = 0x02
	sig[3] = byte(len(rBytes) + rPrefix)
	if rPrefix == 1 {
		sig[4] = 0x00
		copy(sig[5:], rBytes)
	} else {
		copy(sig[4:], rBytes)
	}

	// s
	offset := 4 + len(rBytes) + rPrefix
	sig[offset] = 0x02
	sig[offset+1] = byte(len(sBytes) + sPrefix)
	if sPrefix == 1 {
		sig[offset+2] = 0x00
		copy(sig[offset+3:], sBytes)
	} else {
		copy(sig[offset+2:], sBytes)
	}

	return sig, nil
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
		From:      "A",
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

	// 4. Выводим JSON
	jsonTx, _ := json.MarshalIndent(tx, "", "  ")
	fmt.Printf("\n📤 Transaction JSON:\n%s\n", string(jsonTx))

	// 5. Отправляем
	err = SendTransaction(tx)
	if err != nil {
		panic(err)
	}
}
