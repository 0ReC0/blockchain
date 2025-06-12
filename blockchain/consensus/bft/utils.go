package bft

func HasQuorum(votes map[string][]byte, totalValidators int) bool {
	required := totalValidators*2/3 + 1
	return len(votes) >= required
}
