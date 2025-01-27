package node

import (
	"context"

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

	return &Node{Host: node,
		ctx:              ctx,
		bandwidthCounter: bwctr,
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

// Init initializes the node and its dependencies
// It should be called after the node is created
// and before the node is used
// Init() starts:
// - the mdns discovery service
// - the pubsub service
// - the gossipsub router
// - the relay service
// - the DHT discovery service
func (n *Node) Init() error {

	// start the mdns discovery service
	go discovery.InitMDNSdiscovery(n.ctx, n)
	go discovery.InitDHTDiscovery(n.ctx, n)
	go n.InitStats()

	return nil
}
