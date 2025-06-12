package audit

import (
	"fmt"
	"time"
)

type SecurityEvent struct {
	Timestamp time.Time
	Type      string
	Message   string
	NodeID    string
	Severity  string
}

type AuditLogger struct{}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

func (l *AuditLogger) Log(securityEvent SecurityEvent) {
	fmt.Printf("[%s] [%s] %s - %s\n", securityEvent.Timestamp.Format(time.RFC3339), securityEvent.Severity, securityEvent.Type, securityEvent.Message)
}
