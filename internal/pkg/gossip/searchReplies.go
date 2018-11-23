package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"time"
	"sync"
	"strings"
)

const searchPeriod = 1
const maxMatches = 2

type searchState struct {
	matches  uint8
	keywords []string
	files    map[string]map[uint64][]string
}

var searchStates = struct {
	sync.RWMutex
}{}

var discoveredFiles = make([]files.File, 0)
var matches = struct {
	sync.RWMutex
	l []struct {
		file   string
		chunks map[uint64][]string
	}
}{l: make([]struct {
	file   string
	chunks map[uint64][]string
}, 0)}

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

func (gossiper *Gossiper) initSearchState(keywords []string) {
}

func (gossiper *Gossiper) performPeriodicSearch(
	keywords []string,
	budget uint64,
) {
	// TODO check if state is fine or if budget crossed max, and return
	gossiper.performSearch(gossiper.Name, keywords, budget)
	nextBudget := 2 * budget
	go func() {
		time.Sleep(searchPeriod * time.Second)
		gossiper.performPeriodicSearch(keywords, nextBudget)
	}()
}

func constructSearchIdentifier(keywords []string) string {
	return strings.Join(keywords, ",")
}
