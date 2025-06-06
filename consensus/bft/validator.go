package bft

import (
	"../governance"
	"../pos"
)

type BFTValidator struct {
	*pos.Validator
	Reputation *governance.ReputationSystem
}

func NewBFTValidator(val *pos.Validator) *BFTValidator {
	return &BFTValidator{
		Validator:  val,
		Reputation: governance.NewReputationSystem(),
	}
}

func (v *BFTValidator) EvaluateNode(addr string, score float64) {
	v.Reputation.UpdateReputation(addr, score)
}
