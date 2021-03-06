package p2p

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/otcChain/chord-go/utils"
)

type NetworkV1 struct {
	p2pHost    host.Host
	msgManager *PubSub
	ctxCancel  context.CancelFunc
	ctx        context.Context
}

func newNetwork() *NetworkV1 {
	if _p2pConfig == nil {
		panic("Please init p2p _p2pConfig first")
	}

	opts := _p2pConfig.initOptions()
	ctx, cancel := context.WithCancel(context.Background())

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		panic(err)
	}

	ps, err := newPubSub(ctx, h)
	if err != nil {
		panic(err)
	}
	n := &NetworkV1{
		p2pHost:    h,
		msgManager: ps,
		ctx:        ctx,
		ctxCancel:  cancel,
	}
	n.initRpcApis()

	utils.LogInst().Info().Msgf("p2p with id[%s] created addrs:%s", h.ID(), h.Addrs())
	return n
}

func (nt *NetworkV1) LaunchUp() error {
	return nt.msgManager.start()
}

func (nt *NetworkV1) Destroy() error {
	return nil
}
