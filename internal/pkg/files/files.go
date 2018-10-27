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
	"math"
)

const maxFileChunkSize = 8000
const fileHashSize = 32
const maxChunks = maxFileChunkSize / fileHashSize
const sharedFiles = "client/_SharedFiles/"
const downloads = "client/_Downloads/"
const chunksDownloads = downloads + "/_Chunks/"
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

func ShareFile(
	file bytes.Buffer,
	filename string,
) (*message.FileShareRequest, *message.ValidationResponse) {
	request, err := bufferToShareRequest(file, filename)
	if err != nil {
		return nil, &message.ValidationResponse{Success: false}
	}
	file.Reset()
	response := &message.ValidationResponse{}
	return request, response
}

func bufferToShareRequest(
	buffer bytes.Buffer,
	filename string,
) (*message.FileShareRequest, error) {
	bufferLength := buffer.Len()
	bufferBytes := buffer.Bytes()
	nbChunks :=
		int(math.Ceil(float64(bufferLength) / float64(maxFileChunkSize)))
	if nbChunks > maxChunks {
		return nil, errors.New("file too big")
	}
	chunks := make(map[string][]byte)
	var metafile bytes.Buffer
	for chunk := 0; chunk < nbChunks; chunk++ {
		readChunk := make([]byte, maxFileChunkSize)
		nbBytesRead, _ := buffer.Read(readChunk)
		hash := HashChunk(readChunk[:nbBytesRead])
		_, err := metafile.Write(hash)
		if err != nil {
			return nil, errors.New("error saving file: " + err.Error())
		}
		chunks[HashToKey(hash)] = readChunk[:nbBytesRead]
	}
	err := ioutil.WriteFile(sharedFiles+filename, bufferBytes, os.ModePerm)
	if err != nil {
		fmt.Println("error saving file", err.Error())
		return nil, err
	}
	metahash := HashChunk(metafile.Bytes())
	chunks[HashToKey(metahash)] = metafile.Bytes()
	for chunkName, chunk := range chunks {
		err = ioutil.WriteFile(
			chunksDownloads+chunkName,
			chunk,
			os.ModePerm,
		)
		if err != nil {
			fmt.Println("error saving file", err.Error())
			return nil, err
		}
	}
	return &message.FileShareRequest{
		Name:     filename,
		Size:     uint32(bufferLength),
		Metafile: metafile.Bytes(),
		Metahash: metahash,
	}, nil
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
	chunkeys := make([]string, len(metafile)/fileHashSize)
	for i := 0; i < len(metafile)/fileHashSize; i++ {
		chunkeys[i] = HashToKey(metafile[i*fileHashSize : (i+1)*fileHashSize])
	}
	FileStates.Lock()
	FileStates.m[HashToKey(metafile)].Chunkeys = chunkeys
	FileStates.Unlock()
}

func ShouldIgnoreData(data *message.DataReply) bool {
	if !bytes.Equal(HashChunk(data.Data), data.HashValue) {
		return true
	}
	if IsChunkPresent(data.HashValue) {
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
	return ioutil.ReadFile(chunksDownloads + key)
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

func NextForState(metakey string) ([]byte, error) {
	FileStates.RLock()
	defer FileStates.RUnlock()
	state, hasMetakey := FileStates.m[metakey]
	if !hasMetakey {
		return nil, errors.New("unknown metakey")
	}
	return KeyToHash(state.Chunkeys[state.Index]), nil
}

func IsChunkPresent(key []byte) bool {
	_, err := os.Stat(chunksDownloads + HashToKey(key))
	return err != nil || !os.IsNotExist(err)
}

func DownloadChunk(key []byte, data []byte, sender string) error {
	err := ioutil.WriteFile(chunksDownloads+HashToKey(key), data, os.ModePerm)
	FileStates.RLock()
	fmt.Printf(
		"DOWNLOADING %s chunk %d from %s",
		FileStates.m[HashToKey(key)].Filename,
		FileStates.m[HashToKey(key)].Index+1,
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
		chunk, err := ioutil.ReadFile(chunksDownloads + chunkey)
		if err != nil {
			return err
		}
		file.Write(chunk)
	}
	ioutil.WriteFile(
		downloads+FileStates.m[metakey].Filename,
		file.Bytes(),
		os.ModePerm,
	)
	fmt.Printf("RECONSTRUCTED file %s\n", FileStates.m[metakey].Filename)
	delete(FileStates.m, metakey)
	return nil
}
