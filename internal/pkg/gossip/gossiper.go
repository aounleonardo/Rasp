package gossip

import (
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
	"github.com/dedis/protobuf"
	"strings"
	"time"
)

const maxMsgSize = 1024

var acks Acks

const (
	SEND    = iota
	REQUEST = iota
	NOP     = iota
)

type Gossiper struct {
	Name       string
	uiConn     *net.UDPConn
	uiAddr     *net.UDPAddr
	gossipConn *net.UDPConn
	gossipAddr *net.UDPAddr
	simple     bool
	peers      Peers
	wants      Needs
	rumors     Rumors
	routing    Routes
	privates   Privates
}

func NewGossiper(
	name,
	uiPort,
	gossipAddress string,
	PeerList []string,
	simple bool,
	rtimer int,
) *Gossiper {
	uiAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:"+uiPort)
	uiConn, _ := net.ListenUDP("udp4", uiAddr)

	gossipAddr, _ := net.ResolveUDPAddr("udp4", gossipAddress)
	gossipConn, _ := net.ListenUDP("udp4", gossipAddr)

	peerAddrs := make(map[string]*net.UDPAddr)
	rumors := make(map[string]map[uint32]*message.RumorMessage)
	wants := make(map[string]uint32)
	queue := make(map[string]chan *message.StatusPacket)
	expected := make(map[string]int)

	acks = Acks{queue: queue, expected: expected}

	acks.Lock()
	for _, peer := range PeerList {
		peerAddr, _ := net.ResolveUDPAddr("udp4", peer)
		peerAddrs[peer] = peerAddr
		acks.queue[peer] = make(chan *message.StatusPacket, maxMsgSize)
		acks.expected[peer] = 0
	}
	acks.Unlock()

	gossiper := &Gossiper{
		Name:       name,
		uiConn:     uiConn,
		uiAddr:     uiAddr,
		gossipConn: gossipConn,
		gossipAddr: gossipAddr,
		peers:      Peers{m: peerAddrs},
		simple:     simple,
		rumors:     Rumors{m: rumors},
		wants:      Needs{m: wants},
		routing:    Routes{m: make(map[string]RouteInfo)},
		privates:   Privates{m: make(map[string]*ChatHistory)},
	}

	go gossiper.listenForGossip()
	go gossiper.breakEntropy()

	return gossiper
}

func (gossiper *Gossiper) ListenForClientMessages() {
	for {
		packet := &message.ClientPacket{}
		bytes := make([]byte, maxMsgSize)
		length, sender, err := gossiper.uiConn.ReadFromUDP(bytes)
		if err != nil {
			fmt.Println("Error reading Client Message from UDP: ", err)
			continue
		}
		if length > maxMsgSize {
			fmt.Println(
				"Sent message of size",
				length,
				"is too big, limit is",
				maxMsgSize,
			)
			continue
		}
		protobuf.Decode(bytes, packet)
		gossiper.handleClientPacket(packet, sender)
	}
}

func (gossiper *Gossiper) handleClientPacket(
	packet *message.ClientPacket,
	clientAddr *net.UDPAddr,
) {
	if packet.Rumor != nil {
		gossiper.handleRumorRequest(packet.Rumor, clientAddr)
	} else if packet.Identifier != nil {
		gossiper.handleIdentifierRequest(packet.Identifier, clientAddr)
	} else if packet.Peers != nil {
		gossiper.handlePeersRequest(packet.Peers, clientAddr)
	} else if packet.Messages != nil {
		gossiper.handleMessagesRequest(packet.Messages, clientAddr)
	} else if packet.AddPeer != nil {
		gossiper.handleAddPeersRequest(packet.AddPeer, clientAddr)
	} else if packet.SendPrivate != nil {

	}
}

func (gossiper *Gossiper) buildClientMessage(
	content string,
) *message.GossipPacket {
	if gossiper.simple {
		return &message.GossipPacket{
			Simple: &message.SimpleMessage{
				OriginalName:  gossiper.Name,
				RelayPeerAddr: gossiper.gossipAddr.String(),
				Contents:      content,
			},
		}
	} else {
		return &message.GossipPacket{Rumor: gossiper.createClientRumor(content)}
	}
}

func (gossiper *Gossiper) listenForGossip() {
	for {
		packet := &message.GossipPacket{}
		bytes := make([]byte, maxMsgSize)
		length, sender, err := gossiper.gossipConn.ReadFromUDP(bytes)
		if err != nil {
			fmt.Println("Error reading Peer Message from UDP: ", err)
			continue
		}
		if length > maxMsgSize {
			fmt.Println(
				"Sent message of size",
				length,
				"is too big, limit is",
				maxMsgSize,
			)
			continue
		}
		protobuf.Decode(bytes, packet)
		gossiper.ReceivePacket(packet, sender)
	}
}

func (gossiper *Gossiper) ReceivePacket(
	packet *message.GossipPacket,
	sender *net.UDPAddr,
) {
	gossiper.upsertPeer(sender)
	gossiper.upsertIdentifiers(packet)
	fmt.Printf("PEERS %s\n", gossiper.listPeers())
	if packet.Rumor != nil {
		gossiper.receiveRumorPacket(packet.Rumor, sender)
	} else if packet.Status != nil {
		gossiper.receiveStatusPacket(packet.Status, sender)
	} else if packet.Simple != nil && gossiper.simple {
		gossiper.receiveSimplePacket(packet.Simple, sender)
	} else if packet.Private != nil {
		gossiper.receivePrivateMessage(packet.Private)
	}
}

func (gossiper *Gossiper) upsertPeer(sender *net.UDPAddr) {
	gossiper.peers.Lock()
	defer gossiper.peers.Unlock()
	_, hasPeer := gossiper.peers.m[sender.String()]
	if hasPeer {
		return
	}
	gossiper.peers.m[sender.String()] = sender

	acks.Lock()
	acks.queue[sender.String()] = make(chan *message.StatusPacket, maxMsgSize)
	acks.expected[sender.String()] = 0
	acks.Unlock()
}

func (gossiper *Gossiper) upsertIdentifiers(packet *message.GossipPacket) {
	if packet.Rumor != nil {
		gossiper.upsertOrigin(packet.Rumor.Origin)
	} else if packet.Status != nil {
		for _, peerStatus := range packet.Status.Want {
			gossiper.upsertOrigin(peerStatus.Identifier)
		}
	}
}

func (gossiper *Gossiper) upsertOrigin(origin string) {
	gossiper.rumors.RLock()
	_, hasOrigin := gossiper.rumors.m[origin]
	gossiper.rumors.RUnlock()
	if hasOrigin {
		return
	}
	gossiper.wants.Lock()
	gossiper.rumors.Lock()
	gossiper.rumors.m[origin] = make(map[uint32]*message.RumorMessage)
	gossiper.wants.m[origin] = 1
	gossiper.wants.Unlock()
	gossiper.rumors.Unlock()
}

func (gossiper *Gossiper) listPeers() string {
	keys := make([]string, len(gossiper.peers.m))
	i := 0
	gossiper.peers.RLock()
	for peer := range gossiper.peers.m {
		keys[i] = peer
		i++
	}
	gossiper.peers.RUnlock()
	return strings.Join(keys, ",")
}

func encodeMessage(msg *message.GossipPacket) []byte {
	bytes, err := protobuf.Encode(msg)
	if err != nil {
		fmt.Println("Error encoding gossip packet:", err, "for", msg)
		return nil
	}
	return bytes
}

func (gossiper *Gossiper) breakEntropy() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		_ = <-ticker.C
		neighbour := gossiper.pickRumormongeringPartner(map[string]struct{}{})
		gossiper.sendStatusPacket(neighbour)
	}
}

func (gossiper *Gossiper) nextIdForPeer(identifier string) uint32 {
	gossiper.wants.RLock()
	nextID, ok := gossiper.wants.m[identifier]
	gossiper.wants.RUnlock()
	if !ok {
		return 1
	}
	return nextID
}

func (gossiper *Gossiper) ShutUp() {
	gossiper.gossipConn.Close()
	gossiper.uiConn.Close()
}
