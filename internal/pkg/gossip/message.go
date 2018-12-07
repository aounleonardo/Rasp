package gossip

import (
	"sync"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"time"
	"errors"
	"log"
	"net"
)

type RumorKey struct {
	origin    string
	messageID uint32
}

type MessageKey struct {
	sent      bool
	messageID uint32
}

type Ordering struct {
	sync.RWMutex
	l []RumorKey
}

type OrderedMessages map[uint32]*message.PrivateMessage

type ChatHistory struct {
	sync.RWMutex
	received       OrderedMessages
	sent           OrderedMessages
	nextSend       uint32
	highestReceive uint32
	ordering       []MessageKey
	unordered      []*message.PrivateMessage
}

var messageOrdering Ordering

func wrapAddressAsString(addr *net.UDPAddr) *string {
	if addr == nil {
		return nil
	}
	wrap := addr.String()
	return &wrap
}

func (gossiper *Gossiper) createClientRumor(text string) *message.RumorMessage {
	gossiper.upsertOrigin(gossiper.Name)
	gossiper.wants.RLock()
	id := gossiper.wants.m[gossiper.Name]
	msg := &message.RumorMessage{
		Origin: gossiper.Name,
		ID:     id,
		Text:   text,
	}
	gossiper.wants.RUnlock()

	gossiper.memorizeRumor(msg)
	return msg
}

func (gossiper *Gossiper) memorizeRumor(rumor *message.RumorMessage) {
	gossiper.upsertOrigin(rumor.Origin)

	gossiper.wants.Lock()
	gossiper.wants.m[rumor.Origin] = rumor.ID + 1
	gossiper.wants.Unlock()

	if isRouteRumor(rumor) {
		return
	}

	gossiper.rumors.Lock()
	defer gossiper.rumors.Unlock()
	if _, hasRumor := gossiper.rumors.m[rumor.Origin][rumor.ID]; hasRumor {
		return
	}
	gossiper.rumors.m[rumor.Origin][rumor.ID] = rumor

	messageOrdering.Lock()
	messageOrdering.l = append(
		messageOrdering.l,
		RumorKey{origin: rumor.Origin, messageID: rumor.ID},
	)
	messageOrdering.Unlock()
}

func (gossiper *Gossiper) getMessagesSince(
	startIndex int,
) []message.RumorMessage {
	messageOrdering.RLock()
	gossiper.rumors.RLock()
	defer messageOrdering.RUnlock()
	defer gossiper.rumors.RUnlock()

	if startIndex < 0 || startIndex > len(messageOrdering.l) {
		return nil
	}

	length := len(messageOrdering.l) - startIndex
	messages := make([]message.RumorMessage, length)
	for index := 0; index < length; index++ {
		origin := messageOrdering.l[startIndex+index].origin
		id := messageOrdering.l[startIndex+index].messageID
		messages[index] = *gossiper.rumors.m[origin][id]
	}
	return messages
}

func (gossiper *Gossiper) receivePrivateMessage(
	private *message.PrivateMessage,
) {
	if private.Destination == gossiper.Name {
		fmt.Printf(
			"PRIVATE origin %s hop-limit %d contents %s\n",
			private.Origin,
			private.HopLimit,
			private.Text,
		)
		gossiper.savePrivateMessage(private)
		return
	}
	relayed := *private
	relayed.HopLimit -= 1
	if relayed.HopLimit < 1 {
		return
	}
	gossiper.sendPrivateMessage(&relayed)
}

func (gossiper *Gossiper) sendPrivateMessage(
	private *message.PrivateMessage,
) {
	if private.Origin == gossiper.Name {
		gossiper.savePrivateMessage(private)
	}
	gossiper.relayGossipPacket(
		&message.GossipPacket{Private: private},
		private.Destination,
	)
}

func (gossiper *Gossiper) savePrivateMessage(
	private *message.PrivateMessage,
) {
	sending := true
	peer := private.Destination
	if peer == gossiper.Name {
		peer = private.Origin
		sending = false
	}
	gossiper.upsertChatter(peer)

	gossiper.privates.Lock()
	defer gossiper.privates.Unlock()

	chatHistory := gossiper.privates.m[peer]
	chatHistory.Lock()
	defer chatHistory.Unlock()

	if private.ID == 0 {
		chatHistory.unordered = append(chatHistory.unordered, private)
	} else {
		var m *OrderedMessages
		if sending {
			m = &chatHistory.sent
		} else {
			m = &chatHistory.received
			if private.ID > chatHistory.highestReceive {
				chatHistory.highestReceive = private.ID
			}
		}
		if _, hasMessage := (*m)[private.ID]; hasMessage {
			return
		}
		(*m)[private.ID] = private
		chatHistory.ordering = append(
			chatHistory.ordering,
			MessageKey{sent: sending, messageID: private.ID},
		)
	}
}

func (gossiper *Gossiper) receiveDataRequest(request *message.DataRequest) {
	if request.Destination == gossiper.Name {
		data, err := files.GetChunkForKey(files.HashToKey(request.HashValue))
		if err != nil {
			fmt.Println("error retrieving chunk", err.Error())
			return
		}
		reply := &message.DataReply{
			Origin:      gossiper.Name,
			Destination: request.Origin,
			HopLimit:    hopLimit,
			HashValue:   request.HashValue,
			Data:        data,
		}
		gossiper.relayGossipPacket(
			&message.GossipPacket{DataReply: reply},
			request.Origin,
		)
	}
	relayed := *request
	relayed.HopLimit -= 1
	if relayed.HopLimit < 1 {
		return
	}
	gossiper.relayGossipPacket(
		&message.GossipPacket{DataRequest: &relayed},
		request.Destination,
	)
}

func getSourceOfDataRequest(
	chunk *files.Chunkey,
	defaultPeer *string,
) (string, error) {
	chosen, err := getSourceForChunk(chunk)
	if err != nil && defaultPeer == nil {
		return "", errors.New(fmt.Sprintf(
			"could not find destination for data request %s",
			err.Error(),
		))
	}
	if err != nil {
		return *defaultPeer, nil
	}
	return chosen, nil
}

func (gossiper *Gossiper) sendDataRequest(
	request *message.DataRequest,
	retries int,
) {
	if files.IsChunkPresent(request.HashValue) {
		nextHash, chunkey, _ := files.NextHash(request.HashValue)
		if nextHash != nil {
			destination, err := getSourceOfDataRequest(chunkey, &request.Origin)
			if err != nil {
				log.Fatal("should never happen", err.Error())
			}
			gossiper.sendDataRequest(
				&message.DataRequest{
					Origin:      request.Origin,
					Destination: destination,
					HashValue:   nextHash,
					HopLimit:    hopLimit,
				},
				files.RetryLimit,
			)
		}
		return
	}
	if retries < 0 {
		return
	}
	gossiper.relayGossipPacket(
		&message.GossipPacket{DataRequest: request},
		request.Destination,
	)
	go func() {
		timer := time.NewTimer(5 * time.Second)
		<-timer.C
		gossiper.sendDataRequest(request, retries-1)
	}()
}

func (gossiper *Gossiper) receiveDataReply(reply *message.DataReply) {
	if reply.Destination != gossiper.Name {
		relayed := *reply
		relayed.HopLimit -= 1
		if relayed.HopLimit < 1 {
			return
		}
		gossiper.relayGossipPacket(
			&message.GossipPacket{DataReply: &relayed},
			reply.Destination,
		)
		return
	}
	if files.ShouldIgnoreData(reply) {
		return
	}
	if files.IsAwaitedMetafile(reply.HashValue) {
		files.InitFileState(reply.Data)
	}
	err := files.DownloadChunk(reply.HashValue, reply.Data, reply.Origin)
	if err != nil {
		fmt.Println("error downloading", reply.HashValue, err.Error())
	}
	nextHash, chunkey, err := files.NextHash(reply.HashValue)
	if err != nil {
		return
	}
	if nextHash == nil {
		file, err := files.ReconstructFile(files.HashToKey(reply.HashValue))
		if err != nil {
			fmt.Println(
				"error reconstructing file",
				reply.HashValue,
				err.Error(),
			)
		}
		err = gossiper.saveFile(file)
		if err != nil {
			fmt.Println(
				"error saving file",
				reply.HashValue,
				err.Error(),
			)
		}
		return
	}
	destination, err := getSourceOfDataRequest(chunkey, &reply.Origin)
	if err != nil {
		log.Fatal("should never happen", err.Error())
	}
	gossiper.sendDataRequest(
		&message.DataRequest{
			Origin:      gossiper.Name,
			Destination: destination,
			HopLimit:    hopLimit,
			HashValue:   nextHash,
		},
		files.RetryLimit,
	)
}

func (gossiper *Gossiper) relayGossipPacket(
	packet *message.GossipPacket,
	destination string,
) {
	gossiper.routing.RLock()
	defer gossiper.routing.RUnlock()
	routeInfo, knowsRoute := gossiper.routing.m[destination]
	if !knowsRoute {
		fmt.Println("Doesn't know route to relay gossip packet!", destination)
		return
	}
	bytes := encodeMessage(packet)
	gossiper.gossipConn.WriteToUDP(bytes, routeInfo.nextHop)
}

func (gossiper *Gossiper) upsertChatter(peer string) {
	gossiper.privates.Lock()
	defer gossiper.privates.Unlock()
	if _, hasPeer := gossiper.privates.m[peer]; hasPeer {
		return
	}
	gossiper.privates.m[peer] = &ChatHistory{
		received:       make(map[uint32]*message.PrivateMessage),
		sent:           make(map[uint32]*message.PrivateMessage),
		nextSend:       1,
		highestReceive: 0,
		ordering:       make([]MessageKey, 0),
		unordered:      make([]*message.PrivateMessage, 0),
	}
}

func (gossiper *Gossiper) receiveSearchRequest(request *message.SearchRequest) {
	defer gossiper.timestampRequest(request)
	if gossiper.shouldIgnoreRequest(request) {
		return
	}
	searchReply := &message.SearchReply{
		Origin:      gossiper.Name,
		Destination: request.Origin,
		HopLimit:    hopLimit,
		Results:     gossiper.SearchForKeywords(request.Keywords),
	}
	gossiper.relayGossipPacket(
		&message.GossipPacket{SearchReply: searchReply},
		request.Origin,
	)

	remainingBudget := request.Budget - 1
	if remainingBudget <= 0 {
		return
	}
	gossiper.performSearch(request.Origin, request.Keywords, remainingBudget)
}

func (gossiper *Gossiper) receiveSearchReply(reply *message.SearchReply) {
	gossiper.processSearchResults(reply.Results, reply.Origin)
}
