package files

import (
	"encoding/base64"
	"crypto/sha256"
	"sync"
	"io/ioutil"
	"errors"
	"fmt"
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

func HashToKey(hash []byte) string {
	return base64.URLEncoding.EncodeToString(hash)
}

func KeyToHash(key string) []byte {
	hash, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return hash
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

func GetContainingMetakey(chunkey string) *string {
	FileStates.RLock()
	defer FileStates.RUnlock()
	for metahash, state := range FileStates.m {
		if chunkey == state.Chunkeys[state.Index] {
			return &metahash
		}
	}
	return nil
}

func GetChunkForKey(key string) ([]byte, error) {
	return ioutil.ReadFile(Downloads + key)
}

func NextHash(hashValue []byte) ([]byte, error) {
	metakey := GetContainingMetakey(HashToKey(hashValue))
	if metakey == nil {
		return nil, errors.New("unknown metahash")
	}
	FileStates.RLock()
	defer FileStates.RUnlock()
	if FileStates.m[*metakey].Index+1 >=
		uint32(len(FileStates.m[*metakey].Chunkeys)) {
		return nil, nil
	}
	FileStates.m[*metakey].Index += 1
	return KeyToHash(
		FileStates.m[*metakey].Chunkeys[FileStates.m[*metakey].Index],
	), nil
}
