package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"sync"
	"strings"
	"time"
)

type recentSearches struct {
	sync.RWMutex
	m map[string]time.Time
}
const attentionSpan = 0.5

func (gossiper *Gossiper) distributeBudget(budget uint64) map[string]uint64 {
	gossiper.peers.RLock()
	defer gossiper.peers.RUnlock()
	low := budget / uint64(len(gossiper.peers.m))
	remaining := budget % uint64(len(gossiper.peers.m))
	budgets := make(map[string]uint64)

	i := uint64(0)
	for peer := range gossiper.peers.m {
		if i < remaining {
			budgets[peer] = low + 1
			i++
		} else if low == 0 {
			return budgets
		} else {
			budgets[peer] = low
		}
	}
	return budgets
}

func (gossiper *Gossiper) performSearch(
	origin string,
	keywords []string,
	budget uint64,
) {
	budgets := gossiper.distributeBudget(budget)
	for peer, budget := range budgets {
		gossiper.relayGossipPacket(
			&message.GossipPacket{
				SearchRequest: &message.SearchRequest{
					Origin:   origin,
					Budget:   budget,
					Keywords: keywords,
				},
			},
			peer,
		)
	}
}

func constructRequestIdentifier(request *message.SearchRequest) string {
	return request.Origin + "," + strings.Join(request.Keywords, ",")
}

func (gossiper *Gossiper) shouldIgnoreRequest(
	request *message.SearchRequest,
) bool {
	gossiper.recentSearches.RLock()
	defer gossiper.recentSearches.RUnlock()
	identifier := constructRequestIdentifier(request)
	lastSeen, hasSeen := gossiper.recentSearches.m[identifier]
	return hasSeen && time.Now().Sub(lastSeen).Seconds() < attentionSpan
}

func (gossiper *Gossiper) timestampRequest(request *message.SearchRequest) {
	gossiper.recentSearches.Lock()
	defer gossiper.recentSearches.Unlock()
	identifier := constructRequestIdentifier(request)
	gossiper.recentSearches.m[identifier] = time.Now()
}
