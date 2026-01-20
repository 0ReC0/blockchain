package zkp

import (
	"testing"
)

func TestPrivacyManagerCreation(t *testing.T) {
	_, err := NewPrivacyManager()
	if err != nil {
		t.Errorf("Failed to create PrivacyManager: %v", err)
	}
}
