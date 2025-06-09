package rollup

import (
	"../../../privacy/zkp"
)

type ZKRollup struct {
	ZKP *zkp.ZKPSystem
}

func NewZKRollup() *ZKRollup {
	return &ZKRollup{
		ZKP: zkp.NewZKPSystem(),
	}
}

func (z *ZKRollup) GenerateProof(transactions []string) ([]byte, error) {
	// Генерируем доказательство
	proof, err := z.ZKP.ProveKnowledge([]byte("secret"))
	if err != nil {
		return nil, err
	}
	return proof, nil
}
