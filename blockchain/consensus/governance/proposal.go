// blockchain/consensus/governance/proposal.go
package governance

import (
	"time"
	"blockchain/consensus/pos"
)

// ProposalType определяет тип предложения
type ProposalType string

const (
	ParameterChange ProposalType = "parameter_change"
	ProtocolUpgrade ProposalType = "protocol_upgrade"
	FundingRequest  ProposalType = "funding_request"
)

// Proposal представляет собой предложение для голосования
type Proposal struct {
	ID          string
	Title       string
	Description string
	Author      string
	Type        ProposalType
	Parameters  map[string]interface{} // Параметры для изменения
	Votes       map[string]string      // Адрес -> Выбор
	Threshold   float64                // Минимальный процент голосов для принятия
	StartTime   time.Time
	EndTime     time.Time
	Approved    bool
	Executed    bool
	ValidatorPool *pos.ValidatorPool   // Добавляем пул валидаторов
}