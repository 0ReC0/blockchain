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
	"strconv"
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

// =================== Генерация ключей ===================

func GenerateKeys() (string, string, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	pubKey := &privKey.PublicKey

	// Используем несжатый формат (04 + X + Y)
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()

	xBytesPadded := make([]byte, 32)
	yBytesPadded := make([]byte, 32)
	copy(xBytesPadded[32-len(xBytes):], xBytes)
	copy(yBytesPadded[32-len(yBytes):], yBytes)

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
	type RegisterRequest struct {
		Address string `json:"address"`
		PubKey  string `json:"pubKey"`
	}

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

// =================== Генерация ID ===================

func GenerateSecureToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// =================== HTML UI ===================

// =================== Main ===================

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	// Запускаем веб-сервер
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/addtx", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("📬 /addtx вызван")

		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		// Получаем данные из формы
		var txData struct {
			From      string `json:"From"`
			To        string `json:"To"`
			Amount    string `json:"Amount"`
			IsPrivate string `json:"IsPrivate"`
		}

		if err := json.NewDecoder(r.Body).Decode(&txData); err != nil {
			http.Error(w, "Ошибка парсинга данных", http.StatusBadRequest)
			return
		}

		amount, err := strconv.ParseFloat(txData.Amount, 64)
		if err != nil {
			amount = 10.0
		}

		isPrivate := txData.IsPrivate == "true"

		// Генерируем ключи
		privKey, pubKey, err := GenerateKeys()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Регистрируем публичный ключ
		if err := RegisterPublicKey(txData.From, pubKey); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Создаем транзакцию
		tx := &Transaction{
			ID:        GenerateSecureToken(32),
			From:      txData.From,
			To:        txData.To,
			Amount:    amount,
			Timestamp: time.Now().Unix(),
			IsPrivate: isPrivate,
		}

		// Подписываем
		sig, err := SignTransaction(tx, privKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tx.Signature = sig

		// Отправляем
		if err := SendTransaction(tx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Ответ
		fmt.Fprintf(w, "✅ Транзакция отправлена: %s\n", tx.ID)
	}))

	fmt.Println("🌍 Веб-интерфейс доступен на http://localhost:8000")
	fmt.Println("🔗 Отправка тестовой транзакции: http://localhost:8000/sendtx")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}
