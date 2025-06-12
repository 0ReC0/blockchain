package voting

import "time"

type Proposal struct {
	ID          string
	Title       string
	Description string
	Author      string
	CreatedAt   time.Time
	Votes       map[string]bool // address -> vote (yes/no)
	Quorum      int             // минимальное количество голосов
	Threshold   float64         // минимальный % голосов "за"
	Executed    bool
}
