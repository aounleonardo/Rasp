package main

import (
	"flag"
	"strings"
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"time"
)

const maxMsgSize = 1024

var acks map[string]chan *message.StatusPacket
var expectedAcks map[string]int
var nameToAddr map[string]string

const (
	SEND    = iota
	REQUEST = iota
	NOP     = iota
)

type Gossiper struct {
	ID         uint32
	Name       string
	uiConn     *net.UDPConn
	uiAddr     *net.UDPAddr
	gossipConn *net.UDPConn
	gossipAddr *net.UDPAddr
	simple     bool
	peers      map[string]*net.UDPAddr
	wants      map[string]uint32
	rumors     map[string]map[uint32]*message.RumorMessage
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
	rumors := make(map[string]map[uint32]*message.RumorMessage)
	rumors[gossipAddr.String()] = make(map[uint32]*message.RumorMessage)
	wants := map[string]uint32{gossipAddr.String(): 1}
	nameToAddr = map[string]string{name: gossipAddr.String()}
	acks = make(map[string]chan *message.StatusPacket)
	expectedAcks = make(map[string]int)

	for _, peer := range Peers {
		peerAddr, _ := net.ResolveUDPAddr("udp4", peer)
		peerAddrs[peer] = peerAddr
		rumors[peer] = make(map[uint32]*message.RumorMessage)
		wants[peer] = 1
		acks[peer] = make(chan *message.StatusPacket)
		expectedAcks[peer] = 0
	}

	gossiper := &Gossiper{
		Name:       name,
		uiConn:     uiConn,
		uiAddr:     uiAddr,
		gossipConn: gossipConn,
		gossipAddr: gossipAddr,
		peers:      peerAddrs,
		simple:     simple,
		rumors:     rumors,
		wants:      wants,
	}
	go gossiper.listenForGossip()

	return gossiper
}

func (gossiper *Gossiper) ListenForClientMessages() {
	for {
		packet := &message.ClientPacket{}
		// TODO could make this length an attribute of the Gossiper
		bytes := make([]byte, maxMsgSize)
		length, _, err := gossiper.uiConn.ReadFromUDP(bytes)
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
		fmt.Println("CLIENT MESSAGE", packet.Message)
		fmt.Printf("PEERS %s\n", gossiper.listPeers())
		msg := gossiper.buildClientMessage(packet.Message)

		if gossiper.simple {
			gossiper.forwardSimplePacket(msg, gossiper.gossipAddr)
		} else {
			go gossiper.rumormonger(msg, gossiper.gossipAddr)
		}
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
		id := gossiper.wants[gossiper.gossipAddr.String()]
		msg := &message.RumorMessage{
			Origin: gossiper.Name,
			ID:     id,
			Text:   content,
		}
		gossiper.wants[gossiper.gossipAddr.String()] = id + 1
		gossiper.rumors[gossiper.gossipAddr.String()][id] = msg
		return &message.GossipPacket{Rumor: msg}
	}
}

func (gossiper *Gossiper) listenForGossip() {
	for {
		packet := &message.GossipPacket{}
		bytes := make([]byte, maxMsgSize)
		length, sender, err := gossiper.gossipConn.ReadFromUDP(bytes)
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
		gossiper.ReceiveMessage(packet, sender)
	}
}

func (gossiper *Gossiper) ReceiveMessage(
	packet *message.GossipPacket,
	sender *net.UDPAddr,
) {
	gossiper.upsertPeer(sender)
	fmt.Printf("PEERS %s\n", gossiper.listPeers())
	if packet.Rumor != nil {
		gossiper.receiveRumorPacket(packet, sender)
	} else if packet.Status != nil {
		gossiper.receiveStatusPacket(packet, sender)
	} else if packet.Simple != nil && gossiper.simple {
		gossiper.receiveSimplePacket(packet, sender)
	}
}

func (gossiper *Gossiper) upsertPeer(sender *net.UDPAddr) {
	_, hasPeer := gossiper.peers[sender.String()]
	if hasPeer {
		return
	}
	gossiper.peers[sender.String()] = sender
	gossiper.rumors[sender.String()] = make(map[uint32]*message.RumorMessage)
	gossiper.wants[sender.String()] = 1
	acks[sender.String()] = make(chan *message.StatusPacket)
	expectedAcks[sender.String()] = 0
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

func describeStatusPacket(packet *message.StatusPacket) string {
	ret := ""
	for _, peer := range packet.Want {
		ret += fmt.Sprintf(
			"peer %s nextID %d",
			peer.Identifier,
			peer.NextID,
		)
	}
	return ret
}

func encodeMessage(msg *message.GossipPacket) []byte {
	bytes, err := protobuf.Encode(msg)
	if err != nil {
		fmt.Println("Error encoding gossip packet:", err, "for", msg)
		return nil
	}
	return bytes
}

func (gossiper *Gossiper) receiveSimplePacket(
	packet *message.GossipPacket,
	sender *net.UDPAddr,
) {
	fmt.Printf(
		"SIMPLE MESSAGE origin %s from %s contents %s\n",
		packet.Simple.OriginalName,
		packet.Simple.RelayPeerAddr,
		packet.Simple.Contents,
	)
	gossiper.forwardSimplePacket(packet, sender)
}

func (gossiper *Gossiper) forwardSimplePacket(
	packet *message.GossipPacket,
	sender *net.UDPAddr,
) {
	packet.Simple.RelayPeerAddr = gossiper.gossipAddr.String()
	bytes := encodeMessage(packet)
	if bytes == nil {
		return
	}
	for peer, addr := range gossiper.peers {
		if peer != sender.String() {
			gossiper.gossipConn.WriteToUDP(bytes, addr)
		}
	}
}

func (gossiper *Gossiper) receiveRumorPacket(
	packet *message.GossipPacket,
	sender *net.UDPAddr,
) {
	fmt.Printf(
		"RUMOR origin %s from %s ID %d contents %s\n",
		packet.Rumor.Origin,
		sender.String(),
		packet.Rumor.ID,
		packet.Rumor.Text,
	)

	status := gossiper.constructStatusPacket()
	bytes := encodeMessage(&message.GossipPacket{Status: status})
	gossiper.gossipConn.WriteToUDP(bytes, sender)

	if packet.Rumor.ID != gossiper.nextIdForPeer(packet.Rumor.Origin) {
		return
	}

	gossiper.rumors[packet.Rumor.Origin][packet.Rumor.ID] = packet.Rumor
	gossiper.wants[packet.Rumor.Origin] = packet.Rumor.ID + 1

	go gossiper.rumormonger(packet, sender)
}

func (gossiper *Gossiper) rumormonger(
	msg *message.GossipPacket,
	sender *net.UDPAddr,
) {
	selectedPeerAddr := gossiper.pickRumormongeringPartner(
		map[string]struct{}{sender.String(): {}},
	)
	if selectedPeerAddr == nil {
		return
	}
	gossiper.rumormongerWith(msg, selectedPeerAddr, sender)
}

func (gossiper *Gossiper) pickRumormongeringPartner(
	except map[string]struct{},
) *net.UDPAddr {
	var filteredPeers []string
	for peer := range gossiper.peers {
		_, shouldFilter := except[peer]
		if !shouldFilter {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	if len(filteredPeers) == 0 {
		return nil
	}

	n := rand.Intn(len(filteredPeers))
	return gossiper.peers[filteredPeers[n]]
}

func (gossiper *Gossiper) rumormongerWith(
	msg *message.GossipPacket,
	peer *net.UDPAddr,
	sender *net.UDPAddr,
) {
	bytes := encodeMessage(msg)
	if bytes == nil {
		return
	}
	gossiper.gossipConn.WriteToUDP(bytes, peer)
	expectedAcks[peer.String()] ++

	for {
		var operation int
		var missing message.PeerStatus
		timer := time.NewTimer(time.Second)
		select {
		case ack := <-acks[peer.String()]:
			operation, missing = gossiper.compareStatuses(ack)
		case <-timer.C:
			operation, missing = NOP, message.PeerStatus{}
		}
		switch operation {
		case SEND:
			fmt.Printf("MONGERING with %s\n", peer.String())
			gossiper.sendMissingRumor(&missing, peer)
		case REQUEST:
			status := gossiper.constructStatusPacket()
			packet := &message.GossipPacket{Status: status}
			bytes := encodeMessage(packet)
			if bytes == nil {
				return
			}
			gossiper.gossipConn.WriteToUDP(bytes, peer)
			return
		case NOP:
			if rand.Intn(2) == 0 {

				newPartner := gossiper.pickRumormongeringPartner(
					map[string]struct{}{sender.String(): {}, peer.String(): {}},
				)
				if newPartner != nil {
					fmt.Printf(
						"FLIPPED COIN sending rumor to %s\n",
						newPartner.String(),
					)
					gossiper.rumormonger(msg, sender)
				}
			}
			return
		}
	}
}

func (gossiper *Gossiper) receiveStatusPacket(
	packet *message.GossipPacket,
	sender *net.UDPAddr,
) {
	fmt.Printf(
		"STATUS from %s %s\n",
		sender.String(),
		describeStatusPacket(packet.Status),
	)

	if expectedAcks[sender.String()] > 0 {
		acks[sender.String()] <- packet.Status
		expectedAcks[sender.String()] --
		return
	}
	operation, missing := gossiper.compareStatuses(packet.Status)
	if operation == SEND {
		gossiper.sendMissingRumor(&missing, sender)
	}
	if operation == NOP {
		fmt.Printf("IN SYNC WITH %s\n", sender.String())
	}
}

func (gossiper *Gossiper) constructStatusPacket() *message.StatusPacket {
	peerStatus := make([]message.PeerStatus, len(nameToAddr))
	i := 0
	for name, addr := range nameToAddr {
		peerStatus[i] = message.PeerStatus{
			Identifier: name,
			NextID:     gossiper.wants[addr],
		}
		i++
	}
	return &message.StatusPacket{Want: peerStatus}
}

func (gossiper *Gossiper) compareStatuses(
	packet *message.StatusPacket,
) (int, message.PeerStatus) {
	var needsToRequest = false
	var needed = message.PeerStatus{}
	for _, status := range packet.Want {
		nextID := gossiper.nextIdForPeer(status.Identifier)
		if status.NextID > nextID {
			needsToRequest = true
			needed = message.PeerStatus{
				Identifier: status.Identifier,
				NextID:     nextID,
			}
		} else if status.NextID < nextID {
			return SEND, message.PeerStatus{
				Identifier: status.Identifier,
				NextID:     status.NextID,
			}
		}
	}
	if needsToRequest {
		return REQUEST, needed
	}
	return NOP, needed
}

func (gossiper *Gossiper) sendMissingRumor(
	missing *message.PeerStatus,
	recipient *net.UDPAddr,
) {
	rumor := gossiper.rumors[missing.Identifier][missing.NextID]
	packet := &message.GossipPacket{Rumor: rumor}
	bytes := encodeMessage(packet)
	if bytes == nil {
		return
	}
	gossiper.gossipConn.WriteToUDP(bytes, recipient)
	expectedAcks[recipient.String()] ++
}

// Returns the next id the gossiper wants for a certain peer identifier (name)
// If the gossiper doesn't know yet what address is associated with the name,
// or if it still hasn't received any message from it, it will return 0
func (gossiper *Gossiper) nextIdForPeer(identifier string) uint32 {
	peerAddr, ok := nameToAddr[identifier]
	if !ok {
		return 0
	}
	nextID, ok := gossiper.wants[peerAddr]
	if !ok {
		return 0
	}
	return nextID
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
		"249498",
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

	rand.Seed(time.Now().Unix())

	gossiper := NewGossiper(
		*name,
		*uiPort,
		*gossipAddr,
		peerList,
		*simple,
	)
	defer gossiper.ShutUp()

	gossiper.ListenForClientMessages()
}
