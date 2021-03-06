package main

import (
	"bufio"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"os"
)

var (
	ProtocolID   = protocol.ID("/chord/boot")
	RendezvousID = "block_syncing"
	inputCh      = make(map[string]chan string, 0)
)

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer", err)
			return
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}
	}
}

func writeData(rw *bufio.ReadWriter, ch <-chan string, id string) {
	defer delete(inputCh, id)
	for {
		sendData := <-ch
		_, err := rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer", id, err)
			return
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer", id, err)
			return
		}
	}
}
func handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	in := make(chan string)
	go readData(rw)
	go writeData(rw, in, stream.ID())
	inputCh[stream.ID()] = in

	// 'stream' will stay open until you close it (or the other side closes it).
}
func (bh *BootHost) findPeers() {
	fmt.Println("Searching for other peers...")
	peerChan, err := bh.discovery.FindPeers(bh.ctx, RendezvousID)
	if err != nil {
		panic(err)
	}

	for peerAddr := range peerChan {
		if peerAddr.ID == bh.host.ID() {
			fmt.Println("eee myself...")
			continue
		}
		fmt.Println("Found peer:", peerAddr)
		stream, err := bh.host.NewStream(bh.ctx, peerAddr.ID, ProtocolID)
		if err != nil {
			fmt.Println("Connection failed for peer:", peerAddr)
			continue
		}
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		in := make(chan string)
		go writeData(rw, in, stream.ID())
		go readData(rw)
		inputCh[stream.ID()] = in

		fmt.Println("Connected to:", peerAddr)
	}
}

func (bh *BootHost) chat() {
	bh.host.SetStreamHandler(ProtocolID, handleStream)
	fmt.Println("Announcing ourselves...")
	disc := discovery.NewRoutingDiscovery(bh.dht)
	bh.discovery = disc
	duration, err := disc.Advertise(bh.ctx, RendezvousID)
	if err != nil {
		fmt.Println("advertise self err:", err)
	}

	fmt.Println("Announcing ourselves...", duration)

	bh.findPeers()

	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin", err)
			continue
		}
		if len(inputCh) == 0 {
			fmt.Println("======>there is no stream at all")
			bh.findPeers()
		}
		for _, ch := range inputCh {
			ch <- sendData
		}
	}
}
