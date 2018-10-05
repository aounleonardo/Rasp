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

var acks map[string]chan *message.StatusPacket
var expectedAcks map[string]int

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
	wants      map[string]uint32
	rumors     map[string]map[uint32]*message.RumorMessage
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

	peerAddrs := make(map[string]*net.UDPAddr)
	rumors := make(map[string]map[uint32]*message.RumorMessage)
	rumors[name] = make(map[uint32]*message.RumorMessage)
	wants := map[string]uint32{name: 1}
	acks = make(map[string]chan *message.StatusPacket)
	expectedAcks = make(map[string]int)

	for _, peer := range PeerList {
		peerAddr, _ := net.ResolveUDPAddr("udp4", peer)
		peerAddrs[peer] = peerAddr
		acks[peer] = make(chan *message.StatusPacket)
		expectedAcks[peer] = 0
	}

	gossiper := &Gossiper{
		Name:       name,
		uiConn:     uiConn,
		uiAddr:     uiAddr,
		gossipConn: gossipConn,
		gossipAddr: gossipAddr,
		peers:      Peers{m: peerAddrs},
		simple:     simple,
		rumors:     rumors,
		wants:      wants,
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
		gossiper.handleRumorRequest(packet.Rumor)
	} else if packet.Identifier != nil {
		gossiper.handleIdentifierRequest(packet.Identifier, clientAddr)
	} else if packet.Peers != nil {
		gossiper.handlePeersRequest(packet.Peers, clientAddr)
	} else if packet.Messages != nil {
		gossiper.handleMessagesRequest(packet.Messages, clientAddr)
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
		id := gossiper.wants[gossiper.Name]
		msg := &message.RumorMessage{
			Origin: gossiper.Name,
			ID:     id,
			Text:   content,
		}
		gossiper.wants[gossiper.Name] = id + 1
		gossiper.rumors[gossiper.Name][id] = msg
		return &message.GossipPacket{Rumor: msg}
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
		gossiper.ReceiveMessage(packet, sender)
	}
}

func (gossiper *Gossiper) ReceiveMessage(
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

	acks[sender.String()] = make(chan *message.StatusPacket)
	expectedAcks[sender.String()] = 0
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
	_, hasOrigin := gossiper.rumors[origin]
	if hasOrigin {
		return
	}
	gossiper.rumors[origin] = make(map[uint32]*message.RumorMessage)
	gossiper.wants[origin] = 1
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
	nextID, ok := gossiper.wants[identifier]
	if !ok {
		return 1
	}
	return nextID
}

func (gossiper *Gossiper) ShutUp() {
	gossiper.gossipConn.Close()
	gossiper.uiConn.Close()
}
