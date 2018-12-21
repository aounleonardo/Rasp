package gossip

import (
	"fmt"
	"github.com/aounleonardo/Peerster/internal/pkg/chain"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/dedis/protobuf"
	"net"
)

func (gossiper *Gossiper) handleRumorRequest(
	request *message.RumorRequest,
	clientAddr *net.UDPAddr,
) {
	success := true
	fmt.Println("CLIENT MESSAGE", request.Contents)
	fmt.Printf("PEERS %s\n", gossiper.listPeers())
	msg := gossiper.buildClientMessage(request.Contents)

	if gossiper.simple {
		gossiper.forwardSimplePacket(msg.Simple, gossiper.gossipAddr)
	} else {
		go gossiper.rumormonger(msg.Rumor, gossiper.gossipAddr)
	}
	gossiper.sendValidationToClient(&success, nil, clientAddr)
}

func (gossiper *Gossiper) handleIdentifierRequest(
	clientAddr *net.UDPAddr,
) {
	gossiper.sendToClient(
		&message.IdentifierResponse{Identifier: gossiper.Name},
		clientAddr,
	)
}

func (gossiper *Gossiper) handlePeersRequest(
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
	defer gossiper.sendValidationToClient(&success, nil, clientAddr)
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
	raspMessage *message.RaspMessage,
	clientAddr *net.UDPAddr,
) {
	success := true
	var explanation error
	defer gossiper.sendValidationToClient(&success, &explanation, clientAddr)
	explanation = gossiper.sendPrivateMessage(
		request.Contents,
		request.Destination,
		raspMessage,
	)
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
	defer gossiper.sendValidationToClient(&success, nil, clientAddr)
	if files.IsChunkPresent(request.Metahash) {
		if !files.IsUndergoneMetafile(request.Metahash) {
			success = false
			fmt.Println("metafile present but not undergone")
			return
		}
		chunk, err := files.GetChunkeyForMetakey(metakey)
		if err != nil {
			success = false
			fmt.Println("metafile present but", err.Error())
			return
		}
		destination, err := getSourceOfDataRequest(&chunk, request.Origin)
		if err != nil {
			success = false
			fmt.Println("error getting destination", err.Error())
		}
		err = gossiper.resumeFileDownloadRequest(metakey, destination)
		if err != nil {
			success = false
			fmt.Println("error handling resume request", err.Error())
			return
		}
	} else {
		destination, err := getSourceOfDataRequest(
			&files.Chunkey{Metakey: metakey, Index: uint64(0)},
			request.Origin,
		)
		if err != nil {
			success = false
			fmt.Println("unknown metakey source", err.Error())
			return
		}
		files.NewFileState(metakey, request.Name)
		gossiper.sendDataRequest(
			&message.DataRequest{
				Origin:      gossiper.Name,
				Destination: destination,
				HopLimit:    hopLimit,
				HashValue:   request.Metahash,
			},
			files.RetryLimit,
		)
	}
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

func (gossiper *Gossiper) handlePerformSearchRequest(
	request *message.PerformSearchRequest,
	clientAddr *net.UDPAddr,
) {
	initSearchState(request.Keywords)
	if request.Budget != nil && *request.Budget > 0 {
		gossiper.performSearch(gossiper.Name, request.Keywords, *request.Budget)
	} else {
		gossiper.performPeriodicSearch(request.Keywords, uint64(2))
	}
}

func (gossiper *Gossiper) handleGetSearchesRequest(clientAddr *net.UDPAddr) {
	searches := getAllFileMatches()
	gossiper.sendToClient(
		&message.SearchesResponse{Files: searches},
		clientAddr,
	)
}

func (gossiper *Gossiper) handleCreateMatchRequest(
	request *chain.CreateMatchRequest,
	clientAddr *net.UDPAddr,
) {
	success := true
	var explanation error
	defer gossiper.sendValidationToClient(&success, &explanation, clientAddr)

	raspRequest, err := chain.CreateMatch(
		request.Destination,
		request.Bet,
		request.Move,
		gossiper.Name,
	)
	if err != nil {
		explanation = err
		success = false
		return
	}

	if request.Destination == nil {
		rumour := gossiper.createRaspRumour(raspRequest)
		go gossiper.rumormonger(rumour, gossiper.gossipAddr)
	} else {
		explanation = gossiper.sendPrivateMessage(
			"",
			*request.Destination,
			&message.RaspMessage{Request: raspRequest},
		)
	}
}

func (gossiper *Gossiper) handleAcceptMatchRequest(
	request *chain.AcceptMatchRequest,
	clientAddr *net.UDPAddr,
) {
	success := true
	var explanation error
	defer gossiper.sendValidationToClient(&success, &explanation, clientAddr)

	raspResponse, err := chain.AcceptMatch(
		request.Identifier,
		request.Move,
		gossiper.Name,
	)

	if err != nil {
		explanation = err
		success = false
		return
	}

	explanation = gossiper.sendPrivateMessage(
		"",
		raspResponse.Destination,
		&message.RaspMessage{Response: raspResponse},
	)
}

func (gossiper *Gossiper) handleGetPlayersRequest(
	request *chain.PlayersRequest,
	clientAddr *net.UDPAddr,
) {

	players := &chain.PlayersResponse{}

	chain.GetPlayers(players)

	defer func() {
		gossiper.sendToClient(
			players,
			clientAddr,
		)
	}()

}

func (gossiper *Gossiper) handleGetStatesRequest(
	request *chain.StateRequest,
	clientAddr *net.UDPAddr,
) {

	states := &chain.StateResponse{}

	chain.GetStates(states)

	defer func() {
		gossiper.sendToClient(
			states,
			clientAddr,
		)
	}()

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

func (gossiper *Gossiper) sendValidationToClient(
	success *bool,
	err *error,
	clientAddr *net.UDPAddr,
) {
	if clientAddr == nil {
		return
	}
	explanation := ""
	if err != nil && (*err) != nil {
		explanation = (*err).Error()
	}
	gossiper.sendToClient(
		&message.ValidationResponse{
			Success: *success,
			Error:   explanation,
		},
		clientAddr,
	)
}
