package crosschain

import (
	"time"
)

type CrossChainOracle struct {
	Bridges []*ChainBridge
}

func NewCrossChainOracle() *CrossChainOracle {
	return &CrossChainOracle{
		Bridges: make([]*ChainBridge, 0),
	}
}

func (o *CrossChainOracle) MonitorChains() {
	for {
		for _, bridge := range o.Bridges {
			if len(bridge.Source.Blocks) > 0 {
				latest := bridge.Source.Blocks[len(bridge.Source.Blocks)-1]
				for _, tx := range latest.Transactions {
					if tx.To == "bridge-lock" {
						bridge.MintTokens(tx)
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
