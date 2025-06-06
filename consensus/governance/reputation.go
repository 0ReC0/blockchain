package governance

type Reputation struct {
	Score  float64
	Weight float64
}

type ReputationSystem struct {
	scores map[string]Reputation
}

func NewReputationSystem() *ReputationSystem {
	return &ReputationSystem{
		scores: make(map[string]Reputation),
	}
}

func (r *ReputationSystem) UpdateReputation(node string, score float64) {
	if rep, exists := r.scores[node]; exists {
		rep.Score = (rep.Score + score) / 2
	} else {
		r.scores[node] = Reputation{Score: score}
	}
}

func (r *ReputationSystem) GetReputation(node string) float64 {
	if rep, exists := r.scores[node]; exists {
		return rep.Score
	}
	return 0.5
}
