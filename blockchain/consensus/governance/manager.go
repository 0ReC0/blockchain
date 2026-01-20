// blockchain/consensus/governance/manager.go
package governance

import (
	"fmt"
	"math"
	"time"
	"blockchain/consensus/pos"
)

type GovernanceManager struct {
	Proposals map[string]*Proposal
}

func NewGovernanceManager() *GovernanceManager {
	return &GovernanceManager{
		Proposals: make(map[string]*Proposal),
	}
}

func (g *GovernanceManager) SubmitProposal(p *Proposal) {
	g.Proposals[p.ID] = p
	fmt.Printf("Proposal submitted: %s\n", p.ID)
}

func (g *GovernanceManager) VoteOnProposal(proposalID, voter, choice string) {
	if p, exists := g.Proposals[proposalID]; exists {
		p.Votes[voter] = choice
	}
}

func (g *GovernanceManager) TallyVotes(proposalID string) bool {
	p := g.Proposals[proposalID]
	
	yes := 0
	total := 0
	
	// Получаем список валидаторов из пула
	validators := p.ValidatorPool
	
	// Используем вес голоса, основанный на стейке
	for voter, vote := range p.Votes {
		// Ищем валидатора
		var validator *pos.Validator
		for _, v := range *validators {
			if v.Address == voter {
				validator = v
				break
			}
		}
		
		// Пропускаем, если не нашли валидатора
		if validator == nil {
			continue
		}
		
		// Получаем вес голоса из баланса и комиссий
		weight := validator.Balance + validator.CommissionEarned
		
		if vote == "yes" {
			yes += int(weight)
		}
		total += int(weight)
	}
	
	// Проверяем, есть ли кворум (больше половины валидаторов)
	quorum := len(p.Votes) >= len(*validators)/2
	if !quorum {
		fmt.Printf("Quorum not reached for proposal: %s\n", proposalID)
		return false
	}
	
	// Проверяем, прошло ли предложение порог голосования
	passed := float64(yes)/float64(total) >= p.Threshold
	p.Approved = passed
	
	if passed {
		fmt.Printf("Proposal %s passed with %.2f%% votes\n", proposalID, (float64(yes)/float64(total))*100)
	} else {
		fmt.Printf("Proposal %s failed with %.2f%% votes\n", proposalID, (float64(yes)/float64(total))*100)
	}
	
	return passed
}

// NewProposal создает новое предложение
func NewProposal(id, title, description, author string, proposalType ProposalType, threshold float64, validatorPool *pos.ValidatorPool) *Proposal {
	return &Proposal{
		ID:          id,
		Title:       title,
		Description: description,
		Author:      author,
		Type:        proposalType,
		Parameters:  make(map[string]interface{}),
		Votes:       make(map[string]string),
		Threshold:   threshold,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(7 * 24 * time.Hour), // 7 дней на голосование
		Approved:    false,
		Executed:    false,
		ValidatorPool: validatorPool,
	}
}

// ExecuteProposal выполняняет одобренное предложение
func (g *GovernanceManager) ExecuteProposal(proposalID string) error {
	p := g.Proposals[proposalID]
	
	if !p.Approved {
		return fmt.Errorf("proposal %s not approved", proposalID)
	}
	
	if p.Executed {
		return fmt.Errorf("proposal %s already executed", proposalID)
	}
	
	// Выполняем действия в зависимости от типа предложения
	switch p.Type {
	case ParameterChange:
		// Пример реализации изменения параметров
		for key, value := range p.Parameters {
			fmt.Printf("Changing parameter: %s to %v\n", key, value)
			// Здесь будет реальная логика изменения параметров
		}
		
	case ProtocolUpgrade:
		// Пример реализации обновления протокола
		version, ok := p.Parameters["version"].(string)
		if !ok {
			return fmt.Errorf("invalid version parameter")
		}
		fmt.Printf("Upgrading protocol to version %s\n", version)
		// Здесь будет реальная логика обновления протокола
		
	case FundingRequest:
		// Пример реализации финансирования
		amount, ok := p.Parameters["amount"].(float64)
		if !ok {
			return fmt.Errorf("invalid amount parameter")
		}
		recipient, ok := p.Parameters["recipient"].(string)
		if !ok {
			return fmt.Errorf("invalid recipient parameter")
		}
		fmt.Printf("Funding request approved: %f tokens to %s\n", amount, recipient)
		// Здесь будет реальная логика финансирования
	}
	
	p.Executed = true
	return nil
}