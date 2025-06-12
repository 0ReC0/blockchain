package audit

import (
	"sync"
)

type SecurityAuditor struct {
	Logger *AuditLogger
	Events []SecurityEvent
	mu     sync.Mutex
}

func NewSecurityAuditor() *SecurityAuditor {
	return &SecurityAuditor{
		Logger: NewAuditLogger(),
		Events: make([]SecurityEvent, 0),
	}
}

func (a *SecurityAuditor) RecordEvent(event SecurityEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Events = append(a.Events, event)
	a.Logger.Log(event)
}

func (a *SecurityAuditor) GetEvents() []SecurityEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Events
}
