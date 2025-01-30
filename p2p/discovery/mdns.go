package discovery

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/alejoacosta74/go-logger"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	DefaultRetries = 3

	// DiscoveryInterval is how often we re-publish our mDNS records.
	DiscoveryInterval = time.Hour
)

type MDNSDiscovery struct {
	host   host.Host
	config *DiscoveryConfig
	ctx    context.Context
	// Channel to receive discovered peers from HandlePeerFound
	peerChan chan peer.AddrInfo
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewMDNSDiscovery(h host.Host, config *DiscoveryConfig) *MDNSDiscovery {
	if config == nil {
		config = NewDiscoveryConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &MDNSDiscovery{
		host:     h,
		config:   config,
		peerChan: make(chan peer.AddrInfo, config.MaxPeers),
		ctx:      ctx,
		cancel:   cancel,
		wg:       sync.WaitGroup{},
	}
}

func (d *MDNSDiscovery) Start(ctx context.Context) error {
	discoveryNotifee := &discoveryNotifee{
		h:       d.host,
		ctx:     d.ctx,
		retries: DefaultRetries,
		md:      d,
	}
	discovery := mdns.NewMdnsService(d.host, d.config.ServiceTag, discoveryNotifee)
	return discovery.Start()
}

func (d *MDNSDiscovery) Stop() error {
	d.cancel()
	d.wg.Wait()
	return nil
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h       host.Host
	ctx     context.Context
	retries int // number of retries to connect to a discovered peer
	md      *MDNSDiscovery
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (d *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// in case of panic, recover and log the stack trace
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic in HandlePeerFound", "error", r, "stack", string(debug.Stack()))
		}
	}()
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
			select {
			case d.md.peerChan <- pi:
				logger.Infof("mDNS: discovered peer %s with address %s", pi.ID, pi.Addrs)
			case <-d.ctx.Done():
				logger.Info("context done, stopping discovery")
				return
			default:
				// peer channel is full, discard the peer
				logger.Warn("peer channel is full, discarding peer", "peer", pi.ID)
				return
			}
		} else {
			logger.Debugf("connection attempt %d failed for peer %+v: %v", i+1, pi, err)
		}

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

// DiscoverPeers implements the PeerDiscovery interface by returning the channel
// that receives peers from HandlePeerFound
func (d *MDNSDiscovery) DiscoverPeers(ctx context.Context) (<-chan peer.AddrInfo, error) {
	// Create a new channel for the consumer
	outChan := make(chan peer.AddrInfo, d.config.MaxPeers)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		defer close(outChan)

		for {
			select {
			case peer := <-d.peerChan:
				// Forward discovered peers to the consumer
				select {
				case outChan <- peer:
				case <-ctx.Done():
					return
				case <-d.ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			case <-d.ctx.Done():
				return
			}
		}
	}()

	return outChan, nil
}
