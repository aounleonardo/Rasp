package files

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"io/ioutil"
	"strings"
)

func SearchStatesForKeywords(keywords []string) []*message.SearchResult {
	FileStates.RLock()
	defer FileStates.RUnlock()
	ret := make([]*message.SearchResult, 0)

	for metakey, state := range FileStates.m {
		if HasAnyKeyword(state.Filename, keywords) {
			ret = append(
				ret,
				&message.SearchResult{
					FileName:     state.Filename,
					MetafileHash: KeyToHash(metakey),
					ChunkMap:     BuildChunkmapUpTo(state.Index),
				},
			)
		}
	}

	return ret
}

func getMetafileBytes(metakey string) ([]byte, error) {
	return ioutil.ReadFile(chunksDownloads + metakey)
}

func HasAnyKeyword(filename string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(filename, keyword) {
			return true
		}
	}
	return false
}

func BuildChunkmapUpTo(index uint64) []uint64 {
	chunkmap := make([]uint64, index)
	for i := 1; i <= int(index); i++ {
		chunkmap[i-1] = uint64(i)
	}
	return chunkmap
}
