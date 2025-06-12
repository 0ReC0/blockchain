package bft

import (
	"blockchain/governance/reputation"
	"blockchain/consensus/pos"
)

type BFTValidator struct {
	*pos.Validator
	Reputation *reputation.ReputationSystem
}

func NewBFTValidator(val *pos.Validator) *BFTValidator {
	return &BFTValidator{
		Validator:  val,
		Reputation: reputation.NewReputationSystem(),
	}
}

func (v *BFTValidator) EvaluateNode(addr string, score float64) {
	v.Reputation.UpdateReputation(addr, score)
}
