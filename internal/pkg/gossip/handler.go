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
	gossiper.peers.RLock()
	for peer := range gossiper.peers.m {
		peers = append(peers, peer)
	}
	gossiper.peers.RUnlock()
	gossiper.sendToClient(&message.PeersResponse{Peers: peers}, clientAddr)
}

func (gossiper *Gossiper) handleMessagesRequest(
	request *message.MessagesRequest,
	clientAddr *net.UDPAddr,
) {
	clientStatus := gossiper.buildStatusMap(request.Status.Want)
	gossiper.wants.RLock()
	for peer, wants := range gossiper.wants.m {
		if _, clientHas := clientStatus[peer]; !clientHas && wants > 1 {
			clientStatus[peer] = 1
		}
	}
	gossiper.wants.RUnlock()
	gossiper.sendToClient(
		&message.MessagesResponse{Messages: gossiper.getMessages(clientStatus)},
		clientAddr,
	)
}

func (gossiper *Gossiper) buildStatusMap(
	status []message.PeerStatus,
) map[string]uint32 {
	statusMap := make(map[string]uint32)
	gossiper.wants.RLock()
	for _, peer := range status {
		if _, hasOrigin := gossiper.wants.m[peer.Identifier];
			hasOrigin && peer.NextID > 0 {
			statusMap[peer.Identifier] = peer.NextID
		}
	}
	gossiper.wants.RUnlock()
	return statusMap
}

func (gossiper *Gossiper) getMessages(
	status map[string]uint32,
) []message.PeerMessages {
	peerMessages := make([]message.PeerMessages, len(status))
	i := 0
	for peer, next := range status {
		peerMessages[i] = gossiper.buildPeerMessages(peer, next)
		i++
	}
	return peerMessages
}

func (gossiper *Gossiper) buildPeerMessages(
	peer string,
	start uint32,
) message.PeerMessages {
	gossiper.wants.RLock()
	length := gossiper.wants.m[peer] - start
	if int32(length) < 0 {
		length = 0
	}
	messages := make([]message.RumorMessage, length)
	for i := start; i < gossiper.wants.m[peer]; i++ {
		messages[i-start] = *gossiper.rumors.m[peer][uint32(i)]
	}
	gossiper.wants.RUnlock()
	return message.PeerMessages{Peer: peer, Messages: messages}
}

func (gossiper *Gossiper) sendToClient(
	response interface{},
	clientAddr *net.UDPAddr,
) {
	bytes, err := protobuf.Encode(response)
	if err != nil {
		bytes, _ = protobuf.Encode(err.Error())
		gossiper.uiConn.WriteToUDP(bytes, clientAddr)
		return
	}
	gossiper.uiConn.WriteToUDP(bytes, clientAddr)
}
