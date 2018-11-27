package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"time"
	"sync"
	"strings"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"fmt"
	"errors"
	"math/rand"
	"strconv"
	"net"
)

const searchPeriod = 1 * time.Second
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

func (gossiper *Gossiper) distributeBudget(
	budget uint64,
) map[*net.UDPAddr]uint64 {
	gossiper.peers.RLock()
	defer gossiper.peers.RUnlock()
	low := budget / uint64(len(gossiper.peers.m))
	remaining := budget % uint64(len(gossiper.peers.m))
	budgets := make(map[*net.UDPAddr]uint64)

	i := uint64(0)
	for _, peer := range gossiper.peers.m {
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
	for peerAddr, budget := range budgets {
		bytes := encodeMessage(&message.GossipPacket{
			SearchRequest: &message.SearchRequest{
				Origin: origin,
				Budget: budget,
				Keywords: keywords,
			},
		})
		gossiper.gossipConn.WriteToUDP(bytes, peerAddr)
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
		time.Sleep(searchPeriod)
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
	upsertFileToSearchedStates(result.FileName, metakey)
	if upsertFileToSearchedFiles(result.FileName, metakey) {
		gossiper.sendDataRequest(
			&message.DataRequest{
				Origin:      gossiper.Name,
				Destination: fromPeer,
				HopLimit:    hopLimit,
				HashValue:   result.MetafileHash,
			},
			files.RetryLimit,
		)
		checkNumberOfChunks(metakey, files.RetryLimit)
	}

	searchedFiles.Lock()
	file, _ := searchedFiles.m[metakey]
	for _, index := range result.ChunkMap {
		upsertPeerToChunk(file.chunkDistribution, index, fromPeer)
	}
	probableFileMatched(file)
	searchedFiles.Unlock()

	fmt.Printf(
		"FOUND match %s at %s metafile=%s chunks=%s\n",
		result.FileName,
		fromPeer,
		files.HashToKey(result.MetafileHash),
		chunkmapToString(result.ChunkMap),
	)
}

func chunkmapToString(chunkmap []uint64) string {
	chunkstring := ""
	for _, chunk := range chunkmap {
		chunkstring += strconv.Itoa(int(chunk))
	}
	return chunkstring
}

func probableFileMatched(file *SearchedFile) {
	if file.isMatched {
		return
	}
	if file.totalChunks == nil ||
		uint64(len(file.chunkDistribution)) < *file.totalChunks {
		return
	}

	file.isMatched = true

	searchStates.Lock()
	for searchKey, state := range searchStates.m {
		if stateHasFile(state, file) {
			state.nbMatches++
			if state.nbMatches >= maxMatches {
				fmt.Println("SEARCH FINISHED")
				delete(searchStates.m, searchKey)
			}
		}
	}
	searchStates.Unlock()
}

func stateHasFile(state *SearchState, file *SearchedFile) bool {
	for metakey := range state.files {
		if file.Metakey == metakey {
			return true
		}
	}
	return false
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

func upsertFileToSearchedFiles(filename string, metakey string) bool {
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

func upsertFileToSearchedStates(filename string, metakey string) {
	searchStates.Lock()
	for _, state := range searchStates.m {
		if files.HasAnyKeyword(filename, state.keywords) {
			state.files[metakey] = struct {}{}
		}
	}
	searchStates.Unlock()
}

func checkNumberOfChunks(metakey string, retries int) {
	if retries < 0 {
		return
	}
	searchedFiles.Lock()
	defer searchedFiles.Unlock()
	file := searchedFiles.m[metakey]

	if file.totalChunks != nil {
		probableFileMatched(file)
		return
	}

	if files.IsChunkPresent(files.KeyToHash(metakey)) {
		nbChunks, err := files.GetNumberOfChunksInFile(metakey)
		if err != nil {
			fmt.Println("checkNumberOfChunks error for", metakey, err.Error())
			return
		}
		file.totalChunks = &nbChunks
		probableFileMatched(file)
		return
	}

	go func() {
		time.Sleep(searchPeriod)
		checkNumberOfChunks(metakey, retries-1)
	}()
}

func getSourceForChunk(chunkey files.Chunkey) (string, error) {
	searchedFiles.RLock()
	defer searchedFiles.RUnlock()

	file, hasFile := searchedFiles.m[chunkey.Metakey]
	if !hasFile {
		return "", errors.New(fmt.Sprintf(
			"has not seen file with metakey %s",
			chunkey.Metakey,
		))
	}
	peers, hasChunk := file.chunkDistribution[chunkey.Index]
	if !hasChunk || len(peers) <= 0 {
		return "", errors.New(fmt.Sprintf(
			"has not seen chunk %d for file with metakey %s",
			chunkey.Index,
			chunkey.Metakey,
		))
	}
	n := rand.Intn(len(peers))
	return peers[n], nil
}

func constructSearchIdentifier(keywords []string) string {
	return strings.Join(keywords, ",")
}
