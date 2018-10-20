package gossip

import (
	"sync"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
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
	received  OrderedMessages
	sent      OrderedMessages
	next      uint32
	ordering  []MessageKey
	unordered []*message.PrivateMessage
}

var messageOrdering Ordering

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

	if !isRouteRumor(rumor) {
		gossiper.rumors.Lock()
		messageOrdering.Lock()
		gossiper.rumors.m[rumor.Origin][rumor.ID] = rumor
		messageOrdering.l = append(
			messageOrdering.l,
			RumorKey{origin: rumor.Origin, messageID: rumor.ID},
		)
		gossiper.rumors.Unlock()
		messageOrdering.Unlock()
	}
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
	gossiper.routing.RLock()
	defer gossiper.routing.RUnlock()
	routeInfo, knowsRoute := gossiper.routing.m[private.Destination]
	if !knowsRoute {
		return
	}
	bytes := encodeMessage(&message.GossipPacket{Private: private})
	gossiper.gossipConn.WriteToUDP(bytes, routeInfo.nextHop)
}

func (gossiper *Gossiper) savePrivateMessage(
	private *message.PrivateMessage,
) {
	sending := false
	peer := private.Destination
	if peer == gossiper.Name {
		peer = private.Origin
		sending = true
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

func (gossiper *Gossiper) upsertChatter(peer string) {
	gossiper.privates.Lock()
	defer gossiper.privates.Unlock()
	if _, hasPeer := gossiper.privates.m[peer]; hasPeer {
		return
	}
	gossiper.privates.m[peer] = &ChatHistory{
		received:  make(map[uint32]*message.PrivateMessage),
		sent:      make(map[uint32]*message.PrivateMessage),
		next:      1,
		ordering:  make([]MessageKey, 0),
		unordered: make([]*message.PrivateMessage, 0),
	}
}
