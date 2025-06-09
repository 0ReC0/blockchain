package voting

import (
	"fmt"
	"time"
)

type VotingModule struct {
	Proposals map[string]*Proposal
}

func NewVotingModule() *VotingModule {
	return &VotingModule{
		Proposals: make(map[string]*Proposal),
	}
}

func (v *VotingModule) CreateProposal(title, description, author string, quorum int, threshold float64) string {
	id := GenerateID()
	v.Proposals[id] = &Proposal{
		ID:          id,
		Title:       title,
		Description: description,
		Author:      author,
		CreatedAt:   time.Now(),
		Votes:       make(map[string]bool),
		Quorum:      quorum,
		Threshold:   threshold,
	}
	return id
}

func (v *VotingModule) Vote(proposalID, voter string, approve bool) error {
	proposal, exists := v.Proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found")
	}
	proposal.Votes[voter] = approve
	return nil
}

func (v *VotingModule) IsApproved(proposalID string) (bool, error) {
	proposal, exists := v.Proposals[proposalID]
	if !exists {
		return false, fmt.Errorf("proposal not found")
	}

	yes := 0
	total := len(proposal.Votes)

	for _, vote := range proposal.Votes {
		if vote {
			yes++
		}
	}

	if total < proposal.Quorum {
		return false, fmt.Errorf("quorum not reached")
	}

	approvalRate := float64(yes) / float64(total)
	return approvalRate >= proposal.Threshold, nil
}
