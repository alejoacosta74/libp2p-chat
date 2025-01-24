package app

import (
	"time"

	"github.com/alejoacosta74/go-logger"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func (n *Node) SetupDiscovery() error {
	// setup mDNS discovery to find local peers
	d := mdns.NewMdnsService(n, DiscoveryServiceTag, &discoveryNotifee{h: n})
	return d.Start()
}

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h *Node
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (d *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	logger.Info("discovered peer", "peer", pi.ID)
	err := d.h.Connect(d.h.ctx, pi)
	if err != nil {
		logger.WithFields("peer", pi.ID.String(), "error", err.Error()).Error("failed to connect to peer")
	}
}
