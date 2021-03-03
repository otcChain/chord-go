package p2p

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-pubsub"
	"sync"
)

type PubSub struct {
	ctx  context.Context
	lock sync.Mutex

	dht    *dht.IpfsDHT
	pubSub *pubsub.PubSub
	disc   discovery.Discovery
}

func (s *PubSub) start() error {
	return s.dht.Bootstrap(s.ctx)
}

func newPubSub(ctx context.Context, h host.Host) (*PubSub, error) {
	dhtOpts, err := config.dhtOpts()
	kademliaDHT, err := dht.New(ctx, h, dhtOpts...)
	if err != nil {
		return nil, err
	}
	disc := discovery.NewRoutingDiscovery(kademliaDHT)
	psOption := config.pubSubOpts(disc)
	ps, err := pubsub.NewGossipSub(ctx, h, psOption...)
	if err != nil {
		return nil, err
	}
	return &PubSub{
		ctx:    ctx,
		dht:    kademliaDHT,
		pubSub: ps,
		disc:   disc,
	}, nil
}