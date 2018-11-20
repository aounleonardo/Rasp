package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"errors"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
)

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
