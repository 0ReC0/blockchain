package kyc

import (
	"fmt"
	"time"

	"blockchain/security/audit"
)

type KYCStatus int

const (
	Pending KYCStatus = iota
	Verified
	Rejected
	Suspicious
)

// ComplianceReport представляет отчет о соответствии требованиям
type ComplianceReport struct {
	ReportID             string
	GeneratedAt          time.Time
	TotalUsers           int
	VerifiedUsers        int
	SuspiciousActivities int
	SanctionedAddresses  int
	AMLFindings          []AMLFinding
}

// AMLFinding представляет находку по борьбе с отмыванием денег
type AMLFinding struct {
	UserAddress string
	Activity    string
	Amount      float64
	Timestamp   time.Time
	RiskLevel   string
}

var sanctionedList = map[string]bool{
	"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa": true, // Пример: адрес из санкционного списка
	"UN1CEF5Zk67xxcAAxRE9tjTP8TTs6p8o5":  true, // Еще один пример
}

type User struct {
	Address    string
	FullName   string
	IDNumber   string
	Country    string
	Status     KYCStatus
	RiskScore  float64
	LastUpdate time.Time
}

type KYCManager struct {
	Users   map[string]*User
	Auditor *audit.SecurityAuditor
}

func NewKYCManager(auditor *audit.SecurityAuditor) *KYCManager {
	return &KYCManager{
		Users:   make(map[string]*User),
		Auditor: auditor,
	}
}

func (k *KYCManager) RegisterUser(address, fullName, idNumber, country string) {
	if _, exists := k.Users[address]; exists {
		return
	}
	k.Users[address] = &User{
		Address:    address,
		FullName:   fullName,
		IDNumber:   idNumber,
		Country:    country,
		Status:     Pending,
		RiskScore:  0.5,
		LastUpdate: time.Now(),
	}
	k.Auditor.RecordEvent(audit.SecurityEvent{
		Timestamp: time.Now(),
		Type:      "KYCRegistration",
		Message:   "User registered for KYC: " + address,
		NodeID:    "validator1",
		Severity:  "INFO",
	})
}

func (k *KYCManager) VerifyUser(address string) error {
	user, exists := k.Users[address]
	if !exists {
		return fmt.Errorf("user not found")
	}
	user.Status = Verified
	user.RiskScore = 1.0
	user.LastUpdate = time.Now()
	k.Auditor.RecordEvent(audit.SecurityEvent{
		Timestamp: time.Now(),
		Type:      "KYCVerification",
		Message:   "User verified: " + address,
		NodeID:    "validator1",
		Severity:  "INFO",
	})
	return nil
}

func (k *KYCManager) RejectUser(address string) error {
	user, exists := k.Users[address]
	if !exists {
		return fmt.Errorf("user not found")
	}
	user.Status = Rejected
	user.RiskScore = 0.0
	user.LastUpdate = time.Now()
	k.Auditor.RecordEvent(audit.SecurityEvent{
		Timestamp: time.Now(),
		Type:      "KYCRejection",
		Message:   "User rejected: " + address,
		NodeID:    "validator1",
		Severity:  "WARNING",
	})
	return nil
}

func (k *KYCManager) CheckKYC(address string) (KYCStatus, float64) {
	user, exists := k.Users[address]
	if !exists {
		return Pending, 0
	}
	return user.Status, user.RiskScore
}

func (k *KYCManager) CheckSanctions(address string) bool {
	// Проверяем, есть ли адрес в санкционном списке
	return sanctionedList[address]
}

// ReportSuspiciousActivity регистрирует подозрительную активность
func (k *KYCManager) ReportSuspiciousActivity(address, activity string, amount float64, riskLevel string) {
	// В реальной реализации это должно сохраняться в отдельную структуру данных
	// Для демонстрации просто логируем
	fmt.Printf("🚨 Suspicious activity reported: %s - %s (%.2f) - Risk: %s\n",
		address, activity, amount, riskLevel)

	// Обновляем статус пользователя при необходимости
	if user, exists := k.Users[address]; exists {
		if riskLevel == "HIGH" {
			user.Status = Suspicious
			user.RiskScore = 0.1
		}
	}
}

// GenerateComplianceReport генерирует отчет о соответствии требованиям
func (k *KYCManager) GenerateComplianceReport() *ComplianceReport {
	report := &ComplianceReport{
		ReportID:    fmt.Sprintf("COMPL-%d", time.Now().Unix()),
		GeneratedAt: time.Now(),
		TotalUsers:  len(k.Users),
		AMLFindings: make([]AMLFinding, 0),
	}

	// Подсчитываем верифицированных пользователей
	for _, user := range k.Users {
		if user.Status == Verified {
			report.VerifiedUsers++
		}
	}

	// Подсчитываем санкционированные адреса
	for address := range sanctionedList {
		if _, exists := k.Users[address]; exists {
			report.SanctionedAddresses++
		}
	}

	return report
}

// GetHighRiskUsers возвращает список пользователей с высоким риском
func (k *KYCManager) GetHighRiskUsers() []User {
	var highRiskUsers []User
	for _, user := range k.Users {
		if user.RiskScore < 0.3 || user.Status == Suspicious {
			highRiskUsers = append(highRiskUsers, *user)
		}
	}
	return highRiskUsers
}
