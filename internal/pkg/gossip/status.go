package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
	"strings"
	"net"
)

func describeStatusPacket(packet *message.StatusPacket) string {
	ret := make([]string, len(packet.Want))
	for i, peer := range packet.Want {
		ret[i] = fmt.Sprintf(
			"peer %s nextID %d",
			peer.Identifier,
			peer.NextID,
		)
	}
	return strings.Join(ret, " ")
}

func (gossiper *Gossiper) receiveStatusPacket(
	status *message.StatusPacket,
	sender *net.UDPAddr,
) {
	fmt.Printf(
		"STATUS from %s %s\n",
		sender.String(),
		describeStatusPacket(status),
	)

	if expectedAcks[sender.String()] > 0 {
		acks[sender.String()] <- status
		expectedAcks[sender.String()] --
		return
	}
	operation, missing := gossiper.compareStatuses(status)
	if operation == SEND {
		gossiper.sendMissingRumor(&missing, sender)
	}
	if operation == NOP {
		fmt.Printf("IN SYNC WITH %s\n", sender.String())
	}
}


func (gossiper *Gossiper) sendStatusPacket(to *net.UDPAddr) {
	status := gossiper.constructStatusPacket()
	bytes := encodeMessage(&message.GossipPacket{Status: status})
	if bytes == nil {
		return
	}
	gossiper.gossipConn.WriteToUDP(bytes, to)
}

func (gossiper *Gossiper) constructStatusPacket() *message.StatusPacket {
	gossiper.wants.RLock()
	peerStatus := make([]message.PeerStatus, len(gossiper.wants.m))
	i := 0
	for name, nextID := range gossiper.wants.m {
		peerStatus[i] = message.PeerStatus{
			Identifier: name,
			NextID:     nextID,
		}
		i++
	}
	gossiper.wants.RUnlock()
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
	gossiper.rumors.RLock()
	rumor := gossiper.rumors.m[missing.Identifier][missing.NextID]
	gossiper.rumors.RUnlock()
	packet := &message.GossipPacket{Rumor: rumor}
	bytes := encodeMessage(packet)
	if bytes == nil {
		return
	}
	fmt.Printf("MONGERING with %s\n", recipient.String())
	gossiper.gossipConn.WriteToUDP(bytes, recipient)
	expectedAcks[recipient.String()] ++
}