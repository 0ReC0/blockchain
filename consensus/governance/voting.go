package governance

type Vote struct {
	ProposalID string
	Voter      string
	Choice     string
	Signature  []byte
}

func NewVote(proposalID, voter, choice string, signature []byte) *Vote {
	return &Vote{
		ProposalID: proposalID,
		Voter:      voter,
		Choice:     choice,
		Signature:  signature,
	}
}
