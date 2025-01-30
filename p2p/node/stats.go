package node

import (
	"runtime"
	"time"

	"github.com/alejoacosta74/go-logger"
	"github.com/spf13/viper"
)

const (
	statsInterval = 60 * time.Second
)

func (n *Node) InitStats() {
	roomName := viper.GetString("room")
	topic := "chat-room:" + roomName

	go func() {
		ticker := time.NewTicker(statsInterval)
		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				// bandwidth metrics
				bandwidth := n.bandwidthCounter.GetBandwidthTotals()
				logger.Infof("Rate in: %f, Rate out: %f", bandwidth.RateIn, bandwidth.RateOut)
				// pubsub bandwidth metrics
				pubsubBw := n.bandwidthCounter.GetBandwidthForProtocol("/meshsub/1.1.0")
				logger.Infof("Pubsub Rate in: %f, Rate out: %f", pubsubBw.RateIn, pubsubBw.RateOut)
				// peer metrics
				connectedPeers := len(n.Network().Peers())
				conns := len(n.Network().Conns())
				logger.Infof("Connected peers: %d, Connections: %d", connectedPeers, conns)

				// If using pubsub, log pubsub peers
				if n.PubSub != nil {
					pubsubPeers := len(n.PubSub.ListPeers(topic))
					logger.Infof("Pubsub - Connected peers: %d", pubsubPeers)
				}
				for _, proto := range n.Mux().Protocols() {
					logger.Infof("Active protocol: %s", proto)
				}

				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				logger.Infof("Memory - Alloc: %v MiB, Sys: %v MiB", m.Alloc/1024/1024, m.Sys/1024/1024)
			}
		}
	}()
}
