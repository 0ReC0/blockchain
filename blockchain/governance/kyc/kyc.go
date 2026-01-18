package kyc

import "time"

type KYCStatus int

const (
	Pending KYCStatus = iota
	Verified
	Rejected
	Suspicious
)

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
	Users  map[string]*User
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