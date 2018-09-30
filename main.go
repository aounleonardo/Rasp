package main

import (
	"flag"
	"strings"
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
	"github.com/dedis/protobuf"
)

type Gossiper struct {
	Name    string
	uiConn    *net.UDPConn
	uiAddr *net.UDPAddr
	gossipConn *net.UDPConn
	gossipAddr *net.UDPAddr
	peers map[string]*net.UDPAddr
}

func NewGossiper(
	name,
	uiPort,
	gossipAddress string,
	Peers []string,
	simple bool,
) *Gossiper {
	uiAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:"+uiPort)
	uiConn, _ := net.ListenUDP("udp4", uiAddr)

	gossipAddr, _ := net.ResolveUDPAddr("udp4", gossipAddress)
	gossipConn, _ := net.ListenUDP("udp4", gossipAddr)

	peerAddrs := make(map[string]*net.UDPAddr)
	for _, peer := range Peers {
		peerAddr, _ := net.ResolveUDPAddr("udp4", peer)
		peerAddrs[peer] = peerAddr
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
		fmt.Println("CLIENT MESSAGE", packet.Message)
		fmt.Println(gossiper.listPeers())
		msg := message.SimpleMessage{
			OriginalName: gossiper.Name,
			RelayPeerAddr: gossiper.gossipAddr.String(),
			Contents: packet.Message,
		}
		gossiper.ForwardMessage(
			&message.GossipPacket{Simple: &msg},
			gossiper.gossipAddr,
		)
	}
}

func (gossiper *Gossiper) listenForGossip() {
	for {
		packet := &message.GossipPacket{}
		bytes := make([]byte, 1024)
		length, sender, err := gossiper.gossipConn.ReadFromUDP(bytes)
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
		gossiper.peers[sender.String()] = sender
		fmt.Printf(
			"SIMPLE MESSAGE origin %s from %s contents %s\n",
			packet.Simple.OriginalName,
			packet.Simple.RelayPeerAddr,
			packet.Simple.Contents,
		)
		fmt.Println(gossiper.listPeers())
		gossiper.ForwardMessage(packet, sender)
	}
}

func (gossiper *Gossiper) listPeers() string {
	keys := make([]string, len(gossiper.peers))
	i := 0
	for peer := range gossiper.peers {
		keys[i] = peer
		i++
	}
	return strings.Join(keys, ",")
}

func (gossiper *Gossiper) ForwardMessage(
	msg *message.GossipPacket,
	sender *net.UDPAddr,
) {
	msg.Simple.RelayPeerAddr = gossiper.gossipAddr.String()
	bytes, err := protobuf.Encode(msg)
	if err != nil {
		fmt.Println("Error encoding gossip packet:", err, "for", msg)
		return
	}

	for peer, addr := range gossiper.peers {
		if peer != sender.String() {
			gossiper.gossipConn.WriteToUDP(bytes, addr)
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
