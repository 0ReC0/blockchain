package upgrade

import "time"

type UpgradeStrategy interface {
	ShouldUpgrade() bool
	PlanUpgrade() (*UpgradePlan, error)
	Apply()
}

type LinearUpgradeStrategy struct {
	Manager *UpgradeManager
}

func (s *LinearUpgradeStrategy) ShouldUpgrade() bool {
	return s.Manager.PendingUpgrade != nil && s.Manager.PendingUpgrade.Approved
}

func (s *LinearUpgradeStrategy) PlanUpgrade() (*UpgradePlan, error) {
	return &UpgradePlan{
		Version:     "v2.0.0",
		Description: "Major protocol upgrade",
		ApplyAt:     time.Now().Add(24 * time.Hour),
		Approved:    false,
	}, nil
}

func (s *LinearUpgradeStrategy) Apply() {
	s.Manager.ApplyUpgrade()
}
