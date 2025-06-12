package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
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

type RegisterRequest struct {
	Address string `json:"address"`
	PubKey  string `json:"pubKey"`
}

func GenerateSecureToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// =================== Генерация ключей ===================

func GenerateKeys() (string, string, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	pubKey := &privKey.PublicKey

	// ❌ Используем несжатый формат (04 + X + Y)
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()

	// Делаем длину X и Y равной 32 байтам (для curve P-256)
	xBytesPadded := make([]byte, 32)
	yBytesPadded := make([]byte, 32)
	copy(xBytesPadded[32-len(xBytes):], xBytes)
	copy(yBytesPadded[32-len(yBytes):], yBytes)

	// Формат: 04 || X || Y
	pubKeyBytesUncompressed := append([]byte{0x04}, append(xBytesPadded, yBytesPadded...)...)

	return hex.EncodeToString(privKey.D.Bytes()), hex.EncodeToString(pubKeyBytesUncompressed), nil
}

// =================== Подпись ===================

func CalculateTxHash(tx *Transaction) []byte {
	hash := sha256.Sum256(tx.Serialize())
	return hash[:]
}

func SignTransaction(tx *Transaction, privKeyHex string) (string, error) {
	privKeyBytes, _ := hex.DecodeString(privKeyHex)

	privKey := new(ecdsa.PrivateKey)
	privKey.Curve = elliptic.P256()
	privKey.D = new(big.Int).SetBytes(privKeyBytes)
	privKey.PublicKey.X, privKey.PublicKey.Y = elliptic.P256().ScalarBaseMult(privKeyBytes)

	hash := CalculateTxHash(tx)
	r, s, _ := ecdsa.Sign(rand.Reader, privKey, hash)

	sig, _ := asn1.Marshal(struct {
		R, S *big.Int
	}{R: r, S: s})

	return hex.EncodeToString(sig), nil
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
func RegisterPublicKey(address, pubKey string) error {
	url := "http://localhost:8081/register"

	requestBody := RegisterRequest{
		Address: address,
		PubKey:  pubKey,
	}

	body, _ := json.Marshal(requestBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %d\n%s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("✅ Public key registered. Response: %d\n", resp.StatusCode)
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

	// 2. Регистрируем публичный ключ в блокчейне
	err = RegisterPublicKey("A", pubKey)
	if err != nil {
		panic(err)
	}

	// 3. Создаем транзакцию
	tx := &Transaction{
		ID:        GenerateSecureToken(32),
		From:      "A",
		To:        "validator2",
		Amount:    10.0,
		Timestamp: time.Now().Unix(),
		IsPrivate: false,
	}

	// 4. Подписываем
	sig, err := SignTransaction(tx, privKey)
	if err != nil {
		panic(err)
	}
	tx.Signature = sig

	// 5. Выводим JSON
	jsonTx, _ := json.MarshalIndent(tx, "", "  ")
	fmt.Printf("\n📤 Transaction JSON:\n%s\n", string(jsonTx))

	// 6. Отправляем
	err = SendTransaction(tx)
	if err != nil {
		panic(err)
	}
}
