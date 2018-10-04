package gossip

import (
	"net"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
)

func (gossiper *Gossiper) handleRumorRequest(
	request *message.RumorRequest,
) {
	fmt.Println("CLIENT MESSAGE", request.Contents)
	fmt.Printf("PEERS %s\n", gossiper.listPeers())
	msg := gossiper.buildClientMessage(request.Contents)

	if gossiper.simple {
		gossiper.forwardSimplePacket(msg.Simple, gossiper.gossipAddr)
	} else {
		go gossiper.rumormonger(msg.Rumor, gossiper.gossipAddr)
	}
}

func (gossiper *Gossiper) handleIdentifierRequest(
	request *message.IdentifierRequest,
	clientAddr *net.UDPAddr,
) {
	gossiper.sendToClient(
		&message.IdentifierResponse{Identifier: gossiper.Name},
		clientAddr,
	)
}

func (gossiper *Gossiper) handlePeersRequest(
	request *message.PeersRequest,
	clientAddr *net.UDPAddr,
) {
	var peers []string
	for peer := range gossiper.peers {
		peers = append(peers, peer)
	}
	gossiper.sendToClient(&message.PeersResponse{Peers: peers}, clientAddr)
}

func (gossiper *Gossiper) sendToClient(
	response interface{},
	clientAddr *net.UDPAddr,
) {
	bytes, err := protobuf.Encode(response)
	if err != nil || bytes == nil {
		return
	}
	gossiper.uiConn.WriteToUDP(bytes, clientAddr)
}
