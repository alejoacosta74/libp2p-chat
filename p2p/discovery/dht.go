package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alejoacosta74/go-logger"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

type DHTDiscovery struct {
	host      host.Host
	dht       *dht.IpfsDHT
	discovery *routing.RoutingDiscovery
	config    *DiscoveryConfig
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewDHTDiscovery creates a new DHT-based peer discovery service
func NewDHTDiscovery(h host.Host, config *DiscoveryConfig) *DHTDiscovery {
	if config == nil {
		config = NewDiscoveryConfig()
	}
	return &DHTDiscovery{
		host:   h,
		config: config,
	}
}

// Start initializes the DHT and begins peer discovery
func (d *DHTDiscovery) Start(ctx context.Context) error {
	d.ctx, d.cancel = context.WithCancel(ctx)

	// Initialize the DHT
	kadDHT, err := dht.New(d.ctx, d.host)
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}
	d.dht = kadDHT

	// Bootstrap the DHT
	if err := d.bootstrap(); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Initialize discovery service
	d.discovery = routing.NewRoutingDiscovery(d.dht)

	// Advertise our presence
	util.Advertise(d.ctx, d.discovery, d.config.ServiceTag)

	logger.Info("DHT discovery service started")
	return nil
}

// bootstrap connects to the default bootstrap peers
func (d *DHTDiscovery) bootstrap() error {
	if err := d.dht.Bootstrap(d.ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			logger.Errorf("failed to parse bootstrap peer address: %v", err)
			continue
		}

		wg.Add(1)
		go func(pi *peer.AddrInfo) {
			defer wg.Done()
			if err := d.host.Connect(d.ctx, *pi); err != nil {
				logger.Debugf("failed to connect to bootstrap peer %s: %s", pi.ID, err)
			} else {
				logger.Infof("connected to bootstrap peer: %s", pi.ID)
			}
		}(peerInfo)
	}
	wg.Wait()

	return nil
}

// Stop gracefully shuts down the discovery service
func (d *DHTDiscovery) Stop() error {
	if d.cancel != nil {
		d.cancel()
	}

	// Wait for any ongoing discovery operations to complete
	d.wg.Wait()

	if d.dht != nil {
		if err := d.dht.Close(); err != nil {
			return fmt.Errorf("failed to close DHT: %w", err)
		}
	}

	return nil
}

// DiscoverPeers returns a channel of discovered peer information
func (d *DHTDiscovery) DiscoverPeers(ctx context.Context) (<-chan peer.AddrInfo, error) {
	// Create buffered channel for peer information
	peerChan := make(chan peer.AddrInfo, d.config.MaxPeers)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		defer close(peerChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-d.ctx.Done():
				return
			default:
				peers, err := d.discovery.FindPeers(ctx, d.config.ServiceTag)
				if err != nil {
					logger.Errorf("failed to find peers: %v", err)
					time.Sleep(d.config.RetryTimeout)
					continue
				}

				for p := range peers {
					// Skip self
					if p.ID == d.host.ID() {
						continue
					}

					select {
					case peerChan <- p:
						logger.Debugf("DHT: discovered peer: %s", p.ID)
					case <-ctx.Done():
						return
					case <-d.ctx.Done():
						return
					}
				}

				// Wait before next discovery attempt
				time.Sleep(d.config.RetryTimeout)
			}
		}
	}()

	return peerChan, nil
}
