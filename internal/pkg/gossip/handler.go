package gossip

import (
	"net"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
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

func (gossiper *Gossiper) handleChatsRequest(
	request *message.ChatsRequest,
	clientAddr *net.UDPAddr,
) {
	var chats []string
	gossiper.routing.RLock()
	for origin := range gossiper.routing.m {
		chats = append(chats, origin)
	}
	gossiper.routing.RUnlock()
	gossiper.sendToClient(&message.ChatsResponse{Origins: chats}, clientAddr)
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
	id := chatHistory.nextSend
	chatHistory.nextSend += 1
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

func (gossiper *Gossiper) handleGetPrivateRequest(
	request *message.PrivateGetRequest,
	clientAddr *net.UDPAddr,
) {
	unordered := make([]message.PrivateMessage, 0)
	ordered := make([]message.PrivateMessage, 0)
	unorderedIndex := 1
	orderedIndex := 1
	defer func() {
		gossiper.sendToClient(
			&message.PrivateGetResponse{
				Partner:        request.Partner,
				Unordered:      unordered,
				Ordered:        ordered,
				UnorderedIndex: unorderedIndex,
				OrderedIndex:   orderedIndex,
			},
			clientAddr,
		)
	}()

	gossiper.privates.RLock()
	defer gossiper.privates.RUnlock()
	chatHistory, hasChat := gossiper.privates.m[request.Partner]
	if !hasChat {
		return
	}
	chatHistory.RLock()
	defer chatHistory.RUnlock()

	for i := request.UnorderedIndex; i < len(chatHistory.unordered); i++ {
		unordered = append(unordered, *chatHistory.unordered[i])
	}
	for i := request.OrderedIndex; i < len(chatHistory.ordering); i++ {
		key := chatHistory.ordering[i]
		var private message.PrivateMessage
		if key.sent {
			private = *chatHistory.sent[key.messageID]
		} else {
			private = *chatHistory.received[key.messageID]
		}
		ordered = append(ordered, private)
	}
	unorderedIndex = request.UnorderedIndex
	orderedIndex = request.OrderedIndex
}

func (gossiper *Gossiper) handleFileShareRequest(
	request *message.FileShareRequest,
	clientAddr *net.UDPAddr,
) {
	hashEncoding := files.HashToKey(request.Metahash)
	file := &files.File{
		Name:     request.Name,
		Size:     request.Size,
		Metafile: request.Metafile,
		Metahash: request.Metahash,
	}
	err := gossiper.saveFile(file)
	if err != nil {
		fmt.Println("error saving file", hashEncoding, err.Error())
	}
	gossiper.sendToClient(
		&message.FileShareResponse{
			Name:    request.Name,
			Metakey: hashEncoding,
		},
		clientAddr,
	)
}

func (gossiper *Gossiper) handleFileDownloadRequest(
	request *message.FileDownloadRequest,
	clientAddr *net.UDPAddr,
) {
	metakey := files.HashToKey(request.Metahash)
	success := true
	if files.IsChunkPresent(request.Metahash) {
		err := gossiper.resumeFileDownloadRequest(metakey, request.Origin)
		if err != nil {
			success = false
			return
		}
	} else {
		_ = files.NewFileState(metakey, request.Name)
		gossiper.sendDataRequest(
			&message.DataRequest{
				Origin:      gossiper.Name,
				Destination: request.Origin,
				HopLimit:    hopLimit,
				HashValue:   request.Metahash,
			},
			files.RetryLimit,
		)
	}
	gossiper.sendToClient(
		&message.ValidationResponse{Success: success},
		clientAddr,
	)
}

func (gossiper *Gossiper) resumeFileDownloadRequest(
	metakey string,
	from string,
) error {
	nextHash, err := files.NextForState(metakey)
	if err != nil {
		return err
	}
	gossiper.sendDataRequest(
		&message.DataRequest{
			Origin:      gossiper.Name,
			Destination: from,
			HopLimit:    hopLimit,
			HashValue:   nextHash,
		},
		files.RetryLimit,
	)
	return nil
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
