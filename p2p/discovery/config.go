package discovery

import "time"

// DiscoveryConfig holds common configuration for peer discovery
type DiscoveryConfig struct {
	// ServiceTag is used to identify peers of our application
	ServiceTag string
	// RetryTimeout is how long to wait between discovery attempts
	RetryTimeout time.Duration
	// MaxPeers is the maximum number of peers to discover
	MaxPeers int
}

// NewDiscoveryConfig creates a default discovery configuration
func NewDiscoveryConfig() *DiscoveryConfig {
	return &DiscoveryConfig{
		ServiceTag:   "pubsub-chat-example",
		RetryTimeout: time.Second * 10,
		MaxPeers:     10,
	}
}
