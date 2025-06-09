package upgrade

import (
	"fmt"
	"time"
)

type UpgradePlan struct {
	Version     string
	Description string
	ApplyAt     time.Time
	Approved    bool
}

type UpgradeManager struct {
	CurrentVersion string
	PendingUpgrade *UpgradePlan
	LastUpgrade    time.Time
}

func NewUpgradeManager(currentVersion string) *UpgradeManager {
	return &UpgradeManager{
		CurrentVersion: currentVersion,
		LastUpgrade:    time.Now(),
	}
}

func (u *UpgradeManager) ProposeUpgrade(version, description string, applyAt time.Time) {
	u.PendingUpgrade = &UpgradePlan{
		Version:     version,
		Description: description,
		ApplyAt:     applyAt,
		Approved:    false,
	}
}

func (u *UpgradeManager) ApproveUpgrade() error {
	if u.PendingUpgrade == nil {
		return fmt.Errorf("no pending upgrade")
	}
	u.PendingUpgrade.Approved = true
	return nil
}

func (u *UpgradeManager) ApplyUpgrade() error {
	if u.PendingUpgrade == nil || !u.PendingUpgrade.Approved {
		return fmt.Errorf("upgrade not approved")
	}
	if time.Now().Before(u.PendingUpgrade.ApplyAt) {
		return fmt.Errorf("too early to apply upgrade")
	}
	u.CurrentVersion = u.PendingUpgrade.Version
	u.LastUpgrade = time.Now()
	u.PendingUpgrade = nil
	return nil
}
