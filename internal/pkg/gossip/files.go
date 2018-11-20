package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"errors"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"strings"
	"time"
	"sync"
)

type recentSearches struct {
	sync.RWMutex
	m map[string]time.Time
}
const attentionSpan = 0.5

func (gossiper *Gossiper) saveFile(file *files.File) error {
	gossiper.files.Lock()
	defer gossiper.files.Unlock()

	metakey := files.HashToKey(file.Metahash)
	if _, hasMetakey := gossiper.files.m[metakey];
		hasMetakey {
		return errors.New("file already exists " + metakey)
	}
	gossiper.files.m[metakey] = *file
	return nil
}

func (gossiper *Gossiper) SearchForKeywords(
	keywords []string,
) []*message.SearchResult {
	gossiper.files.RLock()
	defer gossiper.files.RUnlock()
	ret := make([]*message.SearchResult, 0)

	for metakey, file := range gossiper.files.m {
		if files.HasAnyKeyword(file.Name, keywords) {
			ret = append(
				ret,
				&message.SearchResult{
					FileName:     file.Name,
					MetafileHash: files.KeyToHash(metakey),
					ChunkMap: files.BuildChunkmapUpTo(
						file.Size / files.MaxFileChunkSize,
					),
				},
			)
		}
	}

	return append(ret, files.SearchStatesForKeywords(keywords)...)
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