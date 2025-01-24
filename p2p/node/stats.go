package node

import (
	"context"
	"time"

	"github.com/alejoacosta74/go-logger"
)

func (n *Node) InitStats(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logger.Info("Node stats", "ID", n.ID(), "Addresses", n.Addrs(), "Bandwidth", n.bandwidthCounter.GetBandwidthTotals())
			}
		}
	}()
}
