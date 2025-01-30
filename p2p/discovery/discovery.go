package discovery

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerDiscovery defines the interface for peer discovery mechanisms
type PeerDiscovery interface {
	// Start begins the peer discovery process
	Start(context.Context) error
	// Stop halts the peer discovery process
	Stop() error
	// DiscoverPeers returns a channel of discovered peer information
	DiscoverPeers(context.Context) (<-chan peer.AddrInfo, error)
}
