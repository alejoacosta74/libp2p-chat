package app

import (
	"context"

	"github.com/alejoacosta74/go-logger"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type Node struct {
	host.Host
	ctx context.Context
}

func NewNode(ctx context.Context) *Node {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		logger.Fatalf("failed to create host: %v", err)
	}
	return &Node{Host: node, ctx: ctx}
}

// create a new PubSub service using the GossipSub router
func (n *Node) CreatePubSubService() (*pubsub.PubSub, error) {
	return pubsub.NewGossipSub(n.ctx, n)
}
