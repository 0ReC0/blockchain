package layer2

type FraudProof struct {
	BlockHash string
	Proof     []byte
	Validator string
	Timestamp int64
}

func VerifyFraudProof(proof *FraudProof) bool {
	// Пример: простая проверка (в реальности тут будет сложнее)
	if proof.BlockHash == "" || proof.Validator == "" {
		return false
	}
	return true
}
