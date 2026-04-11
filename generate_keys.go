package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Генерация пары ключей ECDSA P-256
func main() {
	fmt.Println("🔑 Генерация пары ключей ECDSA P-256...")
	fmt.Println()

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		return
	}

	pubKey := &privKey.PublicKey

	// Форматируем X и Y координаты (по 32 байта каждая)
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()

	xBytesPadded := make([]byte, 32)
	yBytesPadded := make([]byte, 32)
	copy(xBytesPadded[32-len(xBytes):], xBytes)
	copy(yBytesPadded[32-len(yBytes):], yBytes)

	// Несжатый формат: 04 + X (32 байта) + Y (32 байта) = 65 байт = 130 hex символов
	pubKeyBytes := append([]byte{0x04}, append(xBytesPadded, yBytesPadded...)...)

	fmt.Println("Приватный ключ (hex):")
	fmt.Printf("  %s\n", hex.EncodeToString(privKey.D.Bytes()))
	fmt.Println()
	fmt.Println("Публичный ключ (hex, 130 символов):")
	fmt.Printf("  %s\n", hex.EncodeToString(pubKeyBytes))
	fmt.Println()
	fmt.Println("Длина публичного ключа:", len(hex.EncodeToString(pubKeyBytes)), "символов")
	fmt.Println()
	fmt.Println("💡 Используйте публичный ключ для регистрации:")
	fmt.Printf("  curl -X POST http://localhost:8081/register \\\n")
	fmt.Printf("    -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("    -d '{\"address\":\"user1\",\"pubKey\":\"%s\"}'\n", hex.EncodeToString(pubKeyBytes))
}
