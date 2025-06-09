package fiftyone

import (
	"fmt"
	"time"
)

func (g *FiftyOnePercentGuard) Monitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			if g.IsMajorityAttackPossible() {
				fmt.Println("⚠️ 51% attack risk detected!")
				// Здесь можно вызвать обработчик: alert, slashing, etc.
			}
		}
	}()
}
