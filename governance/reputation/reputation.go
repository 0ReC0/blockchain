package reputation

type Reputation struct {
	Score   float64
	Weight  float64 // влияние на голосование
	History []float64
}

type ReputationModule struct {
	NodeReputation map[string]*Reputation
}

func NewReputationModule() *ReputationModule {
	return &ReputationModule{
		NodeReputation: make(map[string]*Reputation),
	}
}

func (r *ReputationModule) UpdateReputation(node string, score float64) {
	rep, exists := r.NodeReputation[node]
	if !exists {
		rep = &Reputation{
			Score:   100,
			Weight:  1.0,
			History: make([]float64, 0),
		}
		r.NodeReputation[node] = rep
	}
	rep.Score = (rep.Score * 0.8) + (score * 0.2)
	rep.Weight = rep.Score / 100.0
	rep.History = append(rep.History, rep.Score)
}
