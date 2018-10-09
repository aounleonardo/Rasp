package gossip

import (
	"sync"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
)

type RumorKey struct {
	origin    string
	messageID uint32
}

type Ordering struct {
	sync.RWMutex
	l []RumorKey
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
	gossiper.rumors.Lock()
	messageOrdering.Lock()

	if _, hasRumor := gossiper.rumors.m[rumor.Origin][rumor.ID]; !hasRumor {
		gossiper.wants.m[rumor.Origin] = rumor.ID + 1
		gossiper.rumors.m[rumor.Origin][rumor.ID] = rumor
		messageOrdering.l = append(
			messageOrdering.l,
			RumorKey{origin: rumor.Origin, messageID: rumor.ID},
		)
	}

	gossiper.wants.Unlock()
	gossiper.rumors.Unlock()
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
		origin := messageOrdering.l[startIndex + index].origin
		id := messageOrdering.l[startIndex + index].messageID
		messages[index] = *gossiper.rumors.m[origin][id]
	}
	return messages
}
