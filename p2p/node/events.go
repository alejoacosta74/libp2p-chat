package node

import (
	"github.com/alejoacosta74/go-logger"
	"github.com/libp2p/go-libp2p/core/event"
)

// subscribes to the event bus and handles libp2p events as they're received
func (n *Node) eventLoop() {

	// Subscribe to any events of interest
	sub, err := n.Host.EventBus().Subscribe([]interface{}{
		new(event.EvtLocalProtocolsUpdated),
		new(event.EvtLocalAddressesUpdated),
		new(event.EvtLocalReachabilityChanged),
		new(event.EvtNATDeviceTypeChanged),
		new(event.EvtPeerProtocolsUpdated),
		new(event.EvtPeerIdentificationCompleted),
		new(event.EvtPeerIdentificationFailed),
		new(event.EvtPeerConnectednessChanged),
	})
	if err != nil {
		logger.Fatalf("failed to subscribe to peer connectedness events: %s", err)
	}
	defer sub.Close()

	logger.Info("Event listener started")

	for {
		select {
		case evt := <-sub.Out():
			go func(evt interface{}) {
				switch e := evt.(type) {
				case event.EvtLocalProtocolsUpdated:
					logger.Debugf("Event: 'Local protocols updated' - added: %+v, removed: %+v", e.Added, e.Removed)
				case event.EvtLocalAddressesUpdated:
					logger.Debugf("Event: 'Local addresses updated' - added: %+v, removed: %+v", e.Current, e.Removed)
				case event.EvtLocalReachabilityChanged:
					logger.Debugf("Event: 'Local reachability changed': %+v", e.Reachability)
				case event.EvtNATDeviceTypeChanged:
					logger.Debugf("Event: 'NAT device type changed' - DeviceType %v, transport: %v", e.NatDeviceType.String(), e.TransportProtocol.String())
				case event.EvtPeerProtocolsUpdated:
					logger.Debugf("Event: 'Peer protocols updated' - added: %+v, removed: %+v, peer: %+v", e.Added, e.Removed, e.Peer)
				case event.EvtPeerIdentificationCompleted:
					logger.Debugf("Event: 'Peer identification completed' - %v", e.Peer)
				case event.EvtPeerIdentificationFailed:
					logger.Debugf("Event 'Peer identification failed' - peer: %v, reason: %v", e.Peer, e.Reason.Error())
				case event.EvtPeerConnectednessChanged:
					logger.Debugf("Event: 'Peer connectedness change' - Peer %s (peerInfo: %+v) is now %s, protocols: %v, addresses: %v", e.Peer, e.Peer, e.Connectedness)
				case *event.EvtNATDeviceTypeChanged:
					logger.Debugf("Event `NAT device type changed` - DeviceType %v, transport: %v", e.NatDeviceType.String(), e.TransportProtocol.String())
				default:
					logger.Debugf("Received unknown event (type: %T): %+v", e, e)
				}
			}(evt)
		case <-n.ctx.Done():
			logger.Warnf("Context cancel received. Stopping event listener")
			return
		}
	}
}
