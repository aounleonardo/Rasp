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
const chunksDownloads = downloads + "_Chunks/"
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
) (*message.FileShareRequest, *message.FileShareResponse) {
	request, err := bufferToShareRequest(file, filename)
	if err != nil {
		return nil, &message.FileShareResponse{
			Name:    "error sharing: " + err.Error(),
			Metakey: "",
		}
	}
	file.Reset()
	response := &message.FileShareResponse{
		Name:    filename,
		Metakey: HashToKey(request.Metahash),
	}
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
	if FileStates.m == nil {
		FileStates.m = make(map[string]*FileState)
	}
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
	FileStates.m[HashToKey(HashChunk(metafile))].Chunkeys = chunkeys
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

func IsUndergoneMetafile(hash []byte) bool {
	FileStates.RLock()
	_, isUndergoneMetafile := FileStates.m[HashToKey(hash)]
	FileStates.RUnlock()
	return isUndergoneMetafile
}

func IsAwaitedChunk(hash []byte) bool {
	return getContainingMetakey(HashToKey(hash)) != nil
}

func getContainingMetakey(chunkey string) *string {
	FileStates.RLock()
	defer FileStates.RUnlock()
	for metahash, state := range FileStates.m {
		if uint32(len(state.Chunkeys)) >= state.Index && state.Index >= 1 &&
			chunkey == state.Chunkeys[state.Index-1] {
			return &metahash
		}
	}
	return nil
}

func GetChunkForKey(key string) ([]byte, error) {
	return ioutil.ReadFile(chunksDownloads + key)
}

func NextHash(hashValue []byte) ([]byte, error) {
	var metakey string
	if IsUndergoneMetafile(hashValue) {
		metakey = HashToKey(hashValue)
	} else {
		option := getContainingMetakey(HashToKey(hashValue))
		if option == nil {
			return nil, errors.New("unknown metahash")
		}
		metakey = *option
	}
	FileStates.RLock()
	defer FileStates.RUnlock()
	if FileStates.m[metakey].Index >=
		uint32(len(FileStates.m[metakey].Chunkeys)) {
		return nil, nil
	}
	FileStates.m[metakey].Index += 1
	return KeyToHash(
		FileStates.m[metakey].Chunkeys[FileStates.m[metakey].Index-1],
	), nil
}

func NextForState(metakey string) ([]byte, error) {
	FileStates.RLock()
	defer FileStates.RUnlock()
	state, hasMetakey := FileStates.m[metakey]
	if !hasMetakey {
		return nil, errors.New("unknown metakey")
	}
	return KeyToHash(state.Chunkeys[state.Index-1]), nil
}

func IsChunkPresent(key []byte) bool {
	_, err := os.Stat(chunksDownloads + HashToKey(key))
	return err == nil || (err != nil && !os.IsNotExist(err))
}

func DownloadChunk(key []byte, data []byte, sender string) error {
	err := ioutil.WriteFile(chunksDownloads+HashToKey(key), data, os.ModePerm)
	if err != nil {
		fmt.Println("error writing downloaded chunk", err.Error())
		return err
	}
	isMetakey := IsUndergoneMetafile(key)
	if isMetakey {
		return nil
	}
	metakey := getContainingMetakey(HashToKey(key))
	if metakey == nil {
		return errors.New(
			"downloading file corresponding to no metakey" + HashToKey(key),
		)
	}
	FileStates.RLock()
	fmt.Printf(
		"DOWNLOADING %s chunk %d from %s\n",
		FileStates.m[*metakey].Filename,
		FileStates.m[*metakey].Index,
		sender,
	)
	defer FileStates.RUnlock()
	return err
}

func ReconstructFile(key string) error {
	metakey := getContainingMetakey(key)
	if metakey == nil {
		errors.New("reconstructing file corresponding to no metakey" + key)
	}
	FileStates.Lock()
	defer FileStates.Unlock()
	combineChunksIntoFile(
		FileStates.m[*metakey].Chunkeys,
		FileStates.m[*metakey].Filename,
	)
	fmt.Printf("RECONSTRUCTED file %s\n", FileStates.m[*metakey].Filename)
	delete(FileStates.m, *metakey)
	return nil
}

func combineChunksIntoFile(chunkeys []string, filename string) error {
	var file bytes.Buffer
	defer file.Reset()
	for _, chunkey := range chunkeys {
		chunk, err := ioutil.ReadFile(chunksDownloads + chunkey)
		if err != nil {
			return err
		}
		file.Write(chunk)
	}
	return ioutil.WriteFile(
		downloads+filename,
		file.Bytes(),
		os.ModePerm,
	)
}
