// ./contracts/execution/address.go

package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// generateAddress генерирует уникальный адрес на основе временной метки и хеша
func generateAddress() string {
	// Можно использовать временный ID
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	// Хешируем для получения фиксированной длины
	hash := sha256.Sum256([]byte(timestamp))
	return "0x" + hex.EncodeToString(hash[:])[:40] // Ethereum-style 20-byte address
}
