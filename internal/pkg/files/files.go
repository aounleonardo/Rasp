package files

import (
	"encoding/base64"
	"crypto/sha256"
	"sync"
	"io/ioutil"
	"errors"
	"fmt"
	"bytes"
	"os"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
)

const MaxFileChunkSize = 8000
const FileHashSize = 32
const MaxChunks = MaxFileChunkSize / FileHashSize
const SharedFiles = "client/_SharedFiles/"
const Downloads = "client/_Downloads/"
const RetryLimit = 10

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
	Filename string
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

func NewFileState(metahash string, filename string) *FileState {
	newState := &FileState{
		Key:      metahash,
		Chunkeys: nil,
		Index:    0,
		Filename: filename,
	}
	FileStates.Lock()
	FileStates.m[metahash] = newState
	FileStates.Unlock()
	return newState
}

func InitFileState(metafile []byte) {
	chunkeys := make([]string, len(metafile)/FileHashSize)
	for i := 0; i < len(metafile)/FileHashSize; i++ {
		chunkeys[i] = HashToKey(metafile[i*FileHashSize : (i+1)*FileHashSize])
	}
	FileStates.Lock()
	FileStates.m[HashToKey(metafile)].Chunkeys = chunkeys
	FileStates.Unlock()
}

func ShouldIgnoreData(data *message.DataReply) bool {
	if !bytes.Equal(HashChunk(data.Data), data.HashValue) {
		return true
	}
	isAwaitedMetafile := IsAwaitedMetafile(data.HashValue)
	isAwaitedChunk := IsAwaitedChunk(data.HashValue)
	if !isAwaitedMetafile && !isAwaitedChunk {
		return true
	}
	return false
}

func IsAwaitedMetafile(hash []byte) bool {
	FileStates.RLock()
	state, isProcessedMetahash := FileStates.m[HashToKey(hash)]
	isAwaitedMetafile := false
	if isProcessedMetahash && state.Chunkeys == nil {
		isAwaitedMetafile = true
	}
	FileStates.RUnlock()
	return isAwaitedMetafile
}

func IsAwaitedChunk(hash []byte) bool {
	return getContainingMetakey(HashToKey(hash)) != nil
}

func HasMetahashState(key string) bool {
	FileStates.RLock()
	_, isMetahash := FileStates.m[key]
	FileStates.RUnlock()
	return isMetahash
}

func getContainingMetakey(chunkey string) *string {
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
	metakey := getContainingMetakey(HashToKey(hashValue))
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

func IsFilePresent(key []byte) bool {
	_, err := os.Stat(Downloads + HashToKey(key))
	return err != nil || !os.IsNotExist(err)
}

func DownloadChunk(key []byte, data []byte, sender string) error {
	err := ioutil.WriteFile(Downloads + HashToKey(key), data, os.ModePerm)
	FileStates.RLock()
	fmt.Printf(
		"DOWNLOADING %s chunk %d from %s",
		FileStates.m[HashToKey(key)].Filename,
		FileStates.m[HashToKey(key)].Index + 1,
		sender,
	)
	FileStates.RUnlock()
	return err
}

func ReconstructFile(metakey string) error {
	var file bytes.Buffer
	defer file.Reset()
	FileStates.Lock()
	defer FileStates.Unlock()
	for _, chunkey := range FileStates.m[metakey].Chunkeys {
		chunk, err := ioutil.ReadFile(Downloads + chunkey)
		if err != nil {
			return err
		}
		file.Write(chunk)
	}
	ioutil.WriteFile(FileStates.m[metakey].Filename, file.Bytes(), os.ModePerm)
	fmt.Printf("RECONSTRUCTED file %s\n", FileStates.m[metakey].Filename)
	delete(FileStates.m, metakey)
	return nil
}
