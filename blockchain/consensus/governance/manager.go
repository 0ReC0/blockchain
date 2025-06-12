package governance

import (
	"fmt"
)

type GovernanceManager struct {
	Proposals map[string]*Proposal
}

func NewGovernanceManager() *GovernanceManager {
	return &GovernanceManager{
		Proposals: make(map[string]*Proposal),
	}
}

func (g *GovernanceManager) SubmitProposal(p *Proposal) {
	g.Proposals[p.ID] = p
	fmt.Printf("Proposal submitted: %s\n", p.ID)
}

func (g *GovernanceManager) VoteOnProposal(proposalID, voter, choice string) {
	if p, exists := g.Proposals[proposalID]; exists {
		p.Votes[voter] = choice
	}
}

func (g *GovernanceManager) TallyVotes(proposalID string) bool {
	p := g.Proposals[proposalID]
	yes := 0
	total := 0

	for _, vote := range p.Votes {
		if vote == "yes" {
			yes++
		}
		total++
	}

	if float64(yes)/float64(total) >= p.Threshold {
		p.Approved = true
		return true
	}
	return false
}
