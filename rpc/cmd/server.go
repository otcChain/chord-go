package cmd

import (
	"context"
	"fmt"
	"github.com/otcChain/chord-go/p2p"
	"github.com/otcChain/chord-go/pbs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type cmdService struct{}

func (c cmdService) P2PSendTopicMsg(ctx context.Context, msg *pbs.TopicMsg) (*pbs.CommonResponse, error) {

	network, ok := p2p.Inst().(*p2p.NetworkV1)
	if !ok {
		return nil, fmt.Errorf("this test case is not valaible")
	}

	result := network.DebugTopicMsg(msg.Topic, msg.Msg)
	return &pbs.CommonResponse{
		Msg: result,
	}, nil
}

const ServicePort = 8488

func StartCmdService() {
	var address = fmt.Sprintf("127.0.0.1:%d", ServicePort)
	l, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	cmdServer := grpc.NewServer()

	pbs.RegisterCmdServiceServer(cmdServer, &cmdService{})

	reflection.Register(cmdServer)
	if err := cmdServer.Serve(l); err != nil {
		panic(err)
	}
}

func DialToCmdService() pbs.CmdServiceClient {

	var address = fmt.Sprintf("127.0.0.1:%d", ServicePort)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := pbs.NewCmdServiceClient(conn)

	return client
}
