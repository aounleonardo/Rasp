package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"time"
	"sync"
	"strings"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
)

const searchPeriod = 1
const maxMatches = 2
const maxBudget = 32

type SearchState struct {
	nbMatches uint8
	keywords  []string
	files     map[string]struct{}
}

var searchStates = struct {
	sync.RWMutex
	m map[string]*SearchState
}{
	m: make(map[string]*SearchState),
}

type SearchedFile struct {
	Metakey           string
	Name              string
	totalChunks       *uint64
	chunkDistribution map[uint64][]string
	isMatched         bool
}

var searchedFiles = struct {
	sync.RWMutex
	m map[string]*SearchedFile
}{
	m: make(map[string]*SearchedFile),
}

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

func initSearchState(keywords []string) {
	searchStates.Lock()
	defer searchStates.Unlock()
	searchKey := constructSearchIdentifier(keywords)
	if state, hasState := searchStates.m[searchKey]; hasState {
		state.nbMatches = 0
		return
	}
	searchStates.m[searchKey] = &SearchState{
		nbMatches: 0,
		keywords:  keywords,
		files:     make(map[string]struct{}),
	}
}

func (gossiper *Gossiper) performPeriodicSearch(
	keywords []string,
	budget uint64,
) {
	searchKey := constructSearchIdentifier(keywords)
	searchStates.RLock()
	if state, hasState := searchStates.m[searchKey];
		!hasState ||
			budget > maxBudget ||
			state.nbMatches > maxMatches {
		return
	}
	searchStates.RUnlock()

	gossiper.performSearch(gossiper.Name, keywords, budget)
	nextBudget := 2 * budget
	go func() {
		time.Sleep(searchPeriod * time.Second)
		gossiper.performPeriodicSearch(keywords, nextBudget)
	}()
}

func (gossiper *Gossiper) processSearchResults(
	results []*message.SearchResult,
	fromPeer string,
) {
	for _, result := range results {
		gossiper.processResult(result, fromPeer)
	}
}

func (gossiper *Gossiper) processResult(
	result *message.SearchResult,
	fromPeer string,
) {
	metakey := files.HashToKey(result.MetafileHash)
	if upsertSearchedFile(
		result.FileName,
		metakey,
	) {
		gossiper.sendDataRequest(
			&message.DataRequest{
				Origin:      gossiper.Name,
				Destination: fromPeer,
				HopLimit:    hopLimit,
				HashValue:   result.MetafileHash,
			},
			files.RetryLimit,
		)
	}

	searchedFiles.Lock()
	defer searchedFiles.Unlock()
	file, _ := searchedFiles.m[metakey]
	for _, index := range result.ChunkMap {
		upsertPeerToChunk(file.chunkDistribution, index, fromPeer)
	}
	if !file.isMatched && file.totalChunks != nil {
		if uint64(len(file.chunkDistribution)) >= *file.totalChunks {
			newFileMatched(file)
		}
	}
}

func newFileMatched(file *SearchedFile) {
	file.isMatched = true

	searchStates.Lock()
	for searchKey, state := range searchStates.m {
		if files.HasAnyKeyword(file.Name, state.keywords) {
			state.nbMatches++
			if state.nbMatches >= maxMatches {
				delete(searchStates.m, searchKey)
			}
		}
	}
	searchStates.Unlock()
}

// use only if searchedFiles is locked
func upsertPeerToChunk(
	chunkDistribution map[uint64][]string,
	chunk uint64,
	peer string,
) bool {
	if _, hasChunk := chunkDistribution[chunk]; !hasChunk {
		chunkDistribution[chunk] = make([]string, 0)
	}
	if hasPeerInDistribution(chunkDistribution[chunk], peer) {
		return false
	}
	chunkDistribution[chunk] = append(chunkDistribution[chunk], peer)
	return true
}

func hasPeerInDistribution(
	peers []string,
	newPeer string,
) bool {
	for _, peer := range peers {
		if peer == newPeer {
			return true
		}
	}
	return false
}

func upsertSearchedFile(filename string, metakey string) bool {
	searchedFiles.Lock()
	defer searchedFiles.Unlock()
	if _, hasFile := searchedFiles.m[metakey]; !hasFile {
		searchedFiles.m[metakey] = &SearchedFile{
			Metakey:           metakey,
			Name:              filename,
			totalChunks:       nil,
			chunkDistribution: make(map[uint64][]string),
			isMatched:         false,
		}
		return true
	}
	return false
}

func constructSearchIdentifier(keywords []string) string {
	return strings.Join(keywords, ",")
}
