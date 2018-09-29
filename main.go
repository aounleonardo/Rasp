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
	conn    *net.UDPConn
	address *net.UDPAddr
	peers []*net.UDPAddr
}

func NewGossiper(
	name,
	uiPort,
	gossipAddr string,
	PeerList []string,
	simple bool,
) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:"+uiPort)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)

	var peerAddrs []*net.UDPAddr
	for _, peer := range PeerList {
		peerAddr, _ := net.ResolveUDPAddr("udp4", peer)
		peerAddrs = append(peerAddrs, peerAddr)
	}


	gossiper := &Gossiper{
		Name:    name,
		conn:    udpConn,
		address: udpAddr,
		peers: peerAddrs,
	}
	go gossiper.listenForClientMessages()

	SimpleMessages = make(chan message.SimpleMessage)
	go gossiper.ProcessMessages()
	return gossiper
}

func (gossiper *Gossiper) listenForClientMessages() {
	for {
		packet := &message.ClientPacket{}
		// TODO could make this length an attribute of the Gossiper
		bytes := make([]byte, 1024)
		length, _, err := gossiper.conn.ReadFromUDP(bytes)
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
			RelayPeerAddr: gossiper.address.IP.String() + ":" +
				string(gossiper.address.Port),
			Contents: packet.Message,
		}
	}
}

func (gossiper *Gossiper) ProcessMessages() {
	for msg := range SimpleMessages {
		fmt.Println("Received message: ", msg.Contents)
		gossipPacket := &message.GossipPacket{
			Simple: &msg,
		}
		bytes, err := protobuf.Encode(gossipPacket)
		if err != nil {
			fmt.Println("Error encoding gossip packet:", err, "for", msg)
			continue
		}
		for _, peer := range gossiper.peers {
			fmt.Println(peer.String())
			gossiper.conn.WriteToUDP(bytes, peer)
		}
	}
}

func (gossiper *Gossiper) ShutUp() {
	gossiper.conn.Close()
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

	for {
		var input string
		fmt.Scanln(&input)
		if input == "quit" {
			return
		}
	}
}
