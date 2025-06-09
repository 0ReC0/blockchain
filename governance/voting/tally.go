package voting

import "fmt"

func (v *VotingModule) Tally(proposalID string) (yes, no int, err error) {
	proposal, exists := v.Proposals[proposalID]
	if !exists {
		err = fmt.Errorf("proposal not found")
		return
	}
	for _, vote := range proposal.Votes {
		if vote {
			yes++
		} else {
			no++
		}
	}
	return
}
