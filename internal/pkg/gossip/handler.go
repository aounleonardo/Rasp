package gossip

import (
	"net"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
)

func (gossiper *Gossiper) handleRumorRequest(
	request *message.RumorRequest,
	clientAddr *net.UDPAddr,
) {
	fmt.Println("CLIENT MESSAGE", request.Contents)
	fmt.Printf("PEERS %s\n", gossiper.listPeers())
	msg := gossiper.buildClientMessage(request.Contents)

	if gossiper.simple {
		gossiper.forwardSimplePacket(msg.Simple, gossiper.gossipAddr)
	} else {
		go gossiper.rumormonger(msg.Rumor, gossiper.gossipAddr)
	}
	gossiper.sendToClient(&message.ValidationResponse{Success: true}, clientAddr)
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
	messages := gossiper.getMessagesSince(request.StartIndex)
	gossiper.sendToClient(
		&message.MessagesResponse{
			StartIndex: request.StartIndex,
			Messages:   messages,
		},
		clientAddr,
	)
}

func (gossiper *Gossiper) handleAddPeersRequest(
	request *message.AddPeerRequest,
	clientAddr *net.UDPAddr,
) {
	success := true
	defer func() {
		gossiper.sendToClient(
			&message.ValidationResponse{Success: success},
			clientAddr,
		)
	}()
	address, err := net.ResolveUDPAddr(
		"udp4",
		fmt.Sprintf("%s:%s", request.Address, request.Port),
	)

	if err != nil {
		success = false
		return
	}
	gossiper.upsertPeer(address)
}

func (gossiper *Gossiper) handleSendPrivateRequest(
	request *message.PrivatePutRequest,
	clientAddr *net.UDPAddr,
) {
	success := true
	defer func() {
		gossiper.sendToClient(
			&message.ValidationResponse{Success: success},
			clientAddr,
		)
	}()
	gossiper.routing.RLock()
	if _, knowsRoute := gossiper.routing.m[request.Destination]; !knowsRoute {
		success = false
		fmt.Printf(
			"does not know route to destination: %s\n",
			request.Destination,
		)
		return
	}
	gossiper.routing.RUnlock()
	gossiper.upsertChatter(request.Destination)
	gossiper.privates.RLock()
	chatHistory, _ := gossiper.privates.m[request.Destination]
	id := chatHistory.next
	chatHistory.next += 1
	private := &message.PrivateMessage{
		Origin:      gossiper.Name,
		ID:          id,
		Text:        request.Contents,
		Destination: request.Destination,
		HopLimit:    hopLimit,
	}
	gossiper.privates.RUnlock()
	gossiper.receivePrivateMessage(private)
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
