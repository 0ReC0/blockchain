package manager

import (
	"fmt"

	"../governance"
)

type UpgradeManager struct {
	Governance *governance.GovernanceManager
}

func NewUpgradeManager() *UpgradeManager {
	return &UpgradeManager{
		Governance: governance.NewGovernanceManager(),
	}
}

func (u *UpgradeManager) SubmitUpgradeProposal(title, desc, author string) {
	proposal := &governance.Proposal{
		ID:          fmt.Sprintf("upgrade-%d", len(u.Governance.Proposals)+1),
		Title:       title,
		Description: desc,
		Author:      author,
		Votes:       make(map[string]string),
		Threshold:   0.6,
	}
	u.Governance.SubmitProposal(proposal)
}

func (u *UpgradeManager) ApproveUpgrade(proposalID string) {
	if u.Governance.TallyVotes(proposalID) {
		fmt.Printf("Upgrade approved: %s\n", proposalID)
		// Пример обновления консенсуса
		cs := NewConsensusSwitcher(ConsensusBFT)
		cs.StartConsensus()
	}
}
