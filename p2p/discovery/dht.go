package discovery

import (
	"context"
	"fmt"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func dhtStart(ctx context.Context, host host.Host) *dht.IpfsDHT {

	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT.
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			panic(err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerInfo); err != nil {
				fmt.Printf("failed to connect to bootstrap peer %s: %s\n", peerInfo.ID, err)
			} else {
				fmt.Printf("connected to bootstrap peer %s\n", peerInfo.ID)
			}
		}()
	}
	wg.Wait()
	return kademliaDHT
}

func InitDHTDiscovery(ctx context.Context, host host.Host) {
	kademliaDHT := dhtStart(ctx, host)
	routingDiscovery := routing.NewRoutingDiscovery(kademliaDHT)
	util.Advertise(ctx, routingDiscovery, DiscoveryServiceTag)

	anyConnected := false
	for !anyConnected {
		fmt.Println("looking for peers...")
		peersChan, err := routingDiscovery.FindPeers(ctx, DiscoveryServiceTag)
		if err != nil {
			panic(err)
		}
		for peer := range peersChan {
			if peer.ID == host.ID() {
				continue
			}
			// fmt.Println("found peer:", peer)
			if err := host.Connect(ctx, peer); err != nil {
				// fmt.Printf("failed to connect to peer %s: %s\n", peer.ID, err)
			} else {
				fmt.Printf("connected to peer %s\n", peer.ID)
				anyConnected = true
			}
		}
	}
	fmt.Println("peer discovery complete")

}
