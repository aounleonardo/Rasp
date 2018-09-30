package main

import (
	"flag"
	"strings"
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
	"github.com/dedis/protobuf"
)

var SimpleMessages chan message.SimpleMessage

type Gossiper struct {
	Name    string
	uiConn    *net.UDPConn
	uiAddr *net.UDPAddr
	gossipConn *net.UDPConn
	gossipAddr *net.UDPAddr
	peers []*net.UDPAddr
}

func NewGossiper(
	name,
	uiPort,
	gossipAddress string,
	PeerList []string,
	simple bool,
) *Gossiper {
	uiAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:"+uiPort)
	uiConn, _ := net.ListenUDP("udp4", uiAddr)

	gossipAddr, _ := net.ResolveUDPAddr("udp4", gossipAddress)
	gossipConn, _ := net.ListenUDP("udp4", gossipAddr)

	var peerAddrs []*net.UDPAddr
	for _, peer := range PeerList {
		peerAddr, _ := net.ResolveUDPAddr("udp4", peer)
		peerAddrs = append(peerAddrs, peerAddr)
	}


	gossiper := &Gossiper{
		Name:    name,
		uiConn:    uiConn,
		uiAddr: uiAddr,
		gossipConn: gossipConn,
		gossipAddr: gossipAddr,
		peers: peerAddrs,
	}
	go gossiper.listenForGossip()

	SimpleMessages = make(chan message.SimpleMessage)
	go gossiper.ForwardMessages()
	return gossiper
}

func (gossiper *Gossiper) ListenForClientMessages() {
	for {
		packet := &message.ClientPacket{}
		// TODO could make this length an attribute of the Gossiper
		bytes := make([]byte, 1024)
		length, _, err := gossiper.uiConn.ReadFromUDP(bytes)
		if err != nil {
			fmt.Println("Error reading Client Message from UDP: ", err)
			continue
		}
		if length > 1024 {
			fmt.Println(
				"Sent message of size",
				length,
				"is too big, limit is",
				1024,
			)
			continue
		}
		protobuf.Decode(bytes, packet)
		SimpleMessages <- message.SimpleMessage{
			OriginalName: gossiper.Name,
			RelayPeerAddr: gossiper.gossipAddr.String(),
			Contents: packet.Message,
		}
	}
}

func (gossiper *Gossiper) listenForGossip() {
	for {
		packet := &message.GossipPacket{}
		bytes := make([]byte, 1024)
		length, _, err := gossiper.gossipConn.ReadFromUDP(bytes)
		if err != nil {
			fmt.Println("Error reading Client Message from UDP: ", err)
			continue
		}
		if length > 1024 {
			fmt.Println(
				"Sent message of size",
				length,
				"is too big, limit is",
				1024,
			)
			continue
		}
		protobuf.Decode(bytes, packet)
		oldRelay := packet.Simple.RelayPeerAddr
		oldRelayAddr, _ := net.ResolveUDPAddr("udp4", oldRelay)
		gossiper.peers = append(gossiper.peers, oldRelayAddr)
		SimpleMessages <- *packet.Simple
	}
}

func (gossiper *Gossiper) ForwardMessages() {
	for msg := range SimpleMessages {
		relay := msg.RelayPeerAddr
		fmt.Println("Received message: ", msg.Contents)
		gossipPacket := &message.GossipPacket{
			Simple: &msg,
		}
		gossipPacket.Simple.RelayPeerAddr = gossiper.gossipAddr.String()
		bytes, err := protobuf.Encode(gossipPacket)
		if err != nil {
			fmt.Println("Error encoding gossip packet:", err, "for", msg)
			continue
		}
		for _, peer := range gossiper.peers {
			if peer.String() != relay {
				gossiper.gossipConn.WriteToUDP(bytes, peer)
			}
		}
	}
}

func (gossiper *Gossiper) ShutUp() {
	gossiper.gossipConn.Close()
	gossiper.uiConn.Close()
}

func main() {
	var uiPort = flag.String(
		"UIPort",
		"8080",
		"port for the UI client (default \"8080\")",
	)
	var gossipAddr = flag.String(
		"gossipAddr",
		"127.0.0.1:5000",
		"ip:port for the gossiper (default \"127.0.0.1:5000\")",
	)
	var name = flag.String(
		"name",
		"Leo",
		"name of the gossiper",
	)
	var peers = flag.String(
		"peers",
		"",
		"comma separated list of peers of the form ip:port",
	)
	var simple = flag.Bool(
		"simple",
		false,
		"run gossiper in simple broadcast mode",
	)
	flag.Parse()

	var peerList []string
	if len(*peers) > 0 {
		peerList = strings.Split(*peers, ",")
	}

	gossiper := NewGossiper(*name, *uiPort, *gossipAddr, peerList, *simple)
	defer gossiper.ShutUp()

	gossiper.ListenForClientMessages()
}
