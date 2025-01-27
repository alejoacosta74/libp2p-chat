package discovery

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerDiscovery interface {
	Start(context.Context) error
	Stop() error
	DiscoverPeers(context.Context) (<-chan peer.AddrInfo, error)
}
