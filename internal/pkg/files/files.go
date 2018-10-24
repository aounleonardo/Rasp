package files

import (
	"encoding/base64"
	"crypto/sha256"
	"sync"
)

const MaxFileChunkSize = 8000
const FileHashSize = 32
const MaxChunks = MaxFileChunkSize / FileHashSize
const SharedFiles = "client/_SharedFiles/"
const Downloads = "client/_Downloads/"

type File struct {
	Name     string
	Size     uint32
	Metafile []byte
	Metahash []byte
}

type FileState struct {
	Key      string
	Chunkeys []string
	Index    uint32
}

var FileStates struct {
	sync.RWMutex
	m map[string]*FileState
}

func HashToKey(key []byte) string {
	return base64.URLEncoding.EncodeToString(key)
}

func HashChunk(chunk []byte) []byte {
	hasher := sha256.New()
	hasher.Write(chunk)
	return hasher.Sum(nil)
}

func NewFileState(metahash string, metafile []byte) *FileState {
	chunkeys := make([]string, len(metafile)/FileHashSize)
	for i := 0; i < len(metafile)/FileHashSize; i++ {
		chunkeys[i] = HashToKey(metafile[i*FileHashSize : (i+1)*FileHashSize])
	}
	newState := &FileState{Key: metahash, Chunkeys: chunkeys, Index: 0}
	FileStates.Lock()
	FileStates.m[metahash] = newState
	FileStates.Unlock()
	return newState
}

func IsMetahash(key string) bool {
	FileStates.RLock()
	_, isMetahash := FileStates.m[key]
	FileStates.RUnlock()
	return isMetahash
}

func GetContainingMetahash(chunkey string) *string {
	FileStates.RLock()
	defer FileStates.RUnlock()
	for metahash, state := range FileStates.m {
		if chunkey == state.Chunkeys[state.Index] {
			return &metahash
		}
	}
	return nil
}


