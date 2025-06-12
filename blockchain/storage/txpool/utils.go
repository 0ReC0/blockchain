package txpool

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func GenerateID() string {
	data := []byte(time.Now().String() + "salt")
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
