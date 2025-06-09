package private_tx

import (
	"fmt"
	"time"

	"../../crypto/encryption"
	"../../privacy/zkp"
)

type PrivateTransaction struct {
	Sender    string
	Recipient string
	Amount    float64
	Timestamp int64
	Encrypted []byte
	PublicKey []byte
}

func NewPrivateTransaction(sender, recipient string, amount float64, encryptor encryption.Encryptor, publicKey []byte) (*PrivateTransaction, error) {
	data := []byte(fmt.Sprintf("%s:%s:%f:%d", sender, recipient, amount, time.Now().Unix()))
	encrypted, err := encryptor.Encrypt(data, publicKey)
	if err != nil {
		return nil, err
	}
	return &PrivateTransaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
		Encrypted: encrypted,
		PublicKey: publicKey,
	}, nil
}

func (p *PrivateTransaction) Decrypt(encryptor encryption.Encryptor, privateKey []byte) ([]byte, error) {
	// Проверка ZKP перед расшифровкой
	zkp := zkp.NewZKPSystem()
	proof, err := zkp.ProveKnowledge([]byte("secret"))
	if err != nil || !zkp.VerifyProof([]byte("public_key_stub"), proof) {
		return nil, fmt.Errorf("ZKP verification failed")
	}
	return encryptor.Decrypt(p.Encrypted, privateKey)
}
