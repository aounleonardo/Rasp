package gossip

import (
	"fmt"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"net"
)

func (gossiper *Gossiper) receiveSimplePacket(
	msg *message.SimpleMessage,
	sender *net.UDPAddr,
) {
	fmt.Printf(
		"SIMPLE MESSAGE origin %s from %s contents %s\n",
		msg.OriginalName,
		msg.RelayPeerAddr,
		msg.Contents,
	)
	gossiper.forwardSimplePacket(msg, sender)
}

func (gossiper *Gossiper) forwardSimplePacket(
	msg *message.SimpleMessage,
	sender *net.UDPAddr,
) {
	msg.RelayPeerAddr = gossiper.gossipAddr.String()
	bytes := encodeMessage(&message.GossipPacket{Simple: msg})
	if bytes == nil {
		return
	}
	gossiper.peers.RLock()
	for peer, addr := range gossiper.peers.m {
		if peer != sender.String() {
			gossiper.gossipConn.WriteToUDP(bytes, addr)
		}
	}
	gossiper.peers.RUnlock()
}
