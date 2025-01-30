package node

import (
	"context"
	"fmt"

	"github.com/alejoacosta74/go-logger"
	"github.com/alejoacosta74/libp2p-chat-app/p2p/discovery"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pmetrics "github.com/libp2p/go-libp2p/core/metrics"
)

type Node struct {
	host.Host
	ctx              context.Context
	quitCh           chan struct{}
	bandwidthCounter *libp2pmetrics.BandwidthCounter
	discoveries      []discovery.PeerDiscovery
	*pubsub.PubSub
}

func NewNode(ctx context.Context) *Node {
	bwctr := libp2pmetrics.NewBandwidthCounter()
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.BandwidthReporter(bwctr),
		// libp2p.Security(noise.ID, noise.New),
		// libp2p.EnableRelay(),
		// libp2p.NATPortMap(),
	)
	if err != nil {
		logger.Fatalf("failed to create host: %v", err)
	}

	// Create discovery services
	config := discovery.NewDiscoveryConfig()
	dhtDiscovery := discovery.NewDHTDiscovery(node, config)
	mdnsDiscovery := discovery.NewMDNSDiscovery(node, config)

	return &Node{Host: node,
		ctx:              ctx,
		bandwidthCounter: bwctr,
		discoveries:      []discovery.PeerDiscovery{dhtDiscovery, mdnsDiscovery},
	}
}

// create a new PubSub service using the GossipSub router
func (n *Node) CreatePubSubService() (*pubsub.PubSub, error) {
	ps, err := pubsub.NewGossipSub(n.ctx, n)
	if err != nil {
		return nil, err
	}
	n.PubSub = ps
	return ps, nil
}

func (n *Node) Init() error {
	// Start all discovery services
	for _, d := range n.discoveries {
		if err := d.Start(n.ctx); err != nil {
			return fmt.Errorf("failed to start discovery service: %w", err)
		}

		// Start peer discovery in background
		go func(discovery discovery.PeerDiscovery) {
			peerCh, err := discovery.DiscoverPeers(n.ctx)
			if err != nil {
				logger.Errorf("failed to start peer discovery: %v", err)
				return
			}

			for peer := range peerCh {
				if err := n.Connect(n.ctx, peer); err != nil {
					logger.Debugf("failed to connect to peer %s: %s", peer.ID, err)
				}
			}
		}(d)
	}
	go n.eventLoop()
	go n.InitStats()
	return nil
}
