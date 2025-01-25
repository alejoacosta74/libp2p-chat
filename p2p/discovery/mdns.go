package discovery

import (
	"context"
	"time"

	"github.com/alejoacosta74/go-logger"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	DefaultRetries = 3
)

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func InitMDNSdiscovery(ctx context.Context, n host.Host) error {
	// setup mDNS discovery to find local peers
	d := mdns.NewMdnsService(n, DiscoveryServiceTag, &discoveryNotifee{h: n, ctx: ctx, retries: DefaultRetries})
	return d.Start()
}

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h       host.Host
	ctx     context.Context
	retries int // number of retries to connect to a discovered peer
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (d *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// Skip if trying to connect to self
	if pi.ID == d.h.ID() {
		logger.Debug("skipping self connection", "peer", pi.ID)
		return
	}
	logger.Infof("discovered peer %s with address %s", pi.ID, pi.Addrs)

	var err error
	for i := 0; i < d.retries; i++ {
		// Create a context with timeout for the connection attempt
		connectCtx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
		defer cancel()
		logger.Debugf("attempting to connect to peer %s", pi.ID)
		err = d.h.Connect(connectCtx, pi)
		if err == nil {
			logger.Info("successfully connected to peer", "peer", pi.ID)
			return
		}

		logger.Debugf("connection attempt %d failed for peer %s: %s", i+1, pi.ID, err.Error())

		// Wait before retrying
		time.Sleep(time.Second * time.Duration(i+1))
	}

	// If all retries failed, log the error
	if err != nil {
		logger.Error("failed to connect to peer after retries",
			"peer", pi.ID,
			"error", err.Error())
	}
}
