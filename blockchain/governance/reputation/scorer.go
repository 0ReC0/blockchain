package reputation

func (r *ReputationSystem) CalculateScore(node string, success bool) float64 {
	rep, exists := r.NodeReputation[node]
	if !exists {
		return 100
	}
	if success {
		return rep.Score * 1.01 // +1%
	}
	return rep.Score * 0.99 // -1%
}
