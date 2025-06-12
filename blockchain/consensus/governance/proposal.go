package governance

type Proposal struct {
	ID          string
	Title       string
	Description string
	Author      string
	Votes       map[string]string
	Threshold   float64
	Approved    bool
}
