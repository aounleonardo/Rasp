package files

import (
	"crypto/sha256"
	"sync"
	"io/ioutil"
	"errors"
	"fmt"
	"bytes"
	"os"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"math"
	"encoding/hex"
)

const MaxFileChunkSize = 8 * 1024
const fileHashSize = 32
const maxChunks = MaxFileChunkSize / fileHashSize
const SharedFiles = "_SharedFiles/"
const downloads = "_Downloads/"
const chunksDownloads = downloads + "_Chunks/"
const RetryLimit = 10

type File struct {
	Name     string
	Size     uint32
	Metafile []byte
	Metahash []byte
}

func (file File) NbChunks() uint64 {
	return uint64(file.Size/MaxFileChunkSize) + 1
}

type FileState struct {
	Key      string
	Chunkeys []string
	Index    uint64
	Filename string
}

var FileStates struct {
	sync.RWMutex
	m map[string]*FileState
}

type Chunkey struct {
	Metakey string
	Index   uint64
}

func HashToKey(hash []byte) string {
	return hex.EncodeToString(hash)
}

func KeyToHash(key string) []byte {
	hash, err := hex.DecodeString(key)
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
		int(math.Ceil(float64(bufferLength) / float64(MaxFileChunkSize)))
	if nbChunks > maxChunks {
		return nil, errors.New("file too big")
	}
	chunks := make(map[string][]byte)
	var metafile bytes.Buffer
	for chunk := 0; chunk < nbChunks; chunk++ {
		readChunk := make([]byte, MaxFileChunkSize)
		nbBytesRead, _ := buffer.Read(readChunk)
		hash := HashChunk(readChunk[:nbBytesRead])
		_, err := metafile.Write(hash)
		if err != nil {
			return nil, errors.New("error saving file: " + err.Error())
		}
		chunks[HashToKey(hash)] = readChunk[:nbBytesRead]
	}
	err := ioutil.WriteFile(SharedFiles+filename, bufferBytes, os.ModePerm)
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

func NewFileState(metahash string, filename string) {
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
	isAwaitedMetafile := isProcessedMetahash && state.Chunkeys == nil
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
		if uint64(len(state.Chunkeys)) >= state.Index && state.Index >= 1 &&
			chunkey == state.Chunkeys[state.Index-1] {
			return &metahash
		}
	}
	return nil
}

func GetChunkForKey(key string) ([]byte, error) {
	return ioutil.ReadFile(chunksDownloads + key)
}

func NextHash(hashValue []byte) ([]byte, *Chunkey, error) {
	var metakey string
	if IsUndergoneMetafile(hashValue) {
		metakey = HashToKey(hashValue)
	} else {
		option := getContainingMetakey(HashToKey(hashValue))
		if option == nil {
			return nil, nil, errors.New("unknown metahash")
		}
		metakey = *option
	}
	FileStates.RLock()
	defer FileStates.RUnlock()
	if FileStates.m[metakey].Index >=
		uint64(len(FileStates.m[metakey].Chunkeys)) {
		return nil, nil, nil
	}
	FileStates.m[metakey].Index += 1
	return KeyToHash(
		FileStates.m[metakey].Chunkeys[FileStates.m[metakey].Index-1],
	),
		&Chunkey{Metakey: metakey, Index: FileStates.m[metakey].Index},
		nil
}

func GetChunkeyForMetakey(metakey string) (Chunkey, error) {
	FileStates.RLock()
	defer FileStates.RUnlock()
	state, hasState := FileStates.m[metakey]
	if !hasState {
		return Chunkey{}, errors.New(fmt.Sprintf(
			"no state for metakey %s",
			metakey,
		))
	}
	return Chunkey{Metakey: metakey, Index: state.Index + 1}, nil
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

func GetNumberOfChunksInFile(metakey string) (uint64, error) {
	file, err := getMetafileBytes(metakey)
	if err != nil {
		fmt.Println("error while reading metafile", metakey, err.Error())
		return 0, errors.New(fmt.Sprintf(
			"error %s while reading metafile %s",
			err.Error(),
			metakey,
		))
	}
	return uint64(1 + len(file)/MaxFileChunkSize), nil
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

func ReconstructFile(key string) (*File, error) {
	metakey := getContainingMetakey(key)
	if metakey == nil {
		return nil,
			errors.New("reconstructing file corresponding to no metakey" + key)
	}
	FileStates.Lock()
	defer FileStates.Unlock()
	filename := FileStates.m[*metakey].Filename
	size, err := combineChunksIntoFile(
		FileStates.m[*metakey].Chunkeys,
		filename,
	)
	if err != nil {
		return nil,
			errors.New("reconstructing file failed for metakey" + key)
	}
	fmt.Printf("RECONSTRUCTED file %s\n", FileStates.m[*metakey].Filename)
	delete(FileStates.m, *metakey)
	metafile, err := getMetafileBytes(key)
	if err != nil {
		return nil,
			errors.New("could not find metafile for metakey" + key)
	}
	return &File{
		Name:     filename,
		Size:     size,
		Metafile: metafile,
		Metahash: KeyToHash(*metakey),
	},
		nil
}

func combineChunksIntoFile(chunkeys []string, filename string) (uint32, error) {
	var file bytes.Buffer
	defer file.Reset()
	for _, chunkey := range chunkeys {
		chunk, err := ioutil.ReadFile(chunksDownloads + chunkey)
		if err != nil {
			return 0, err
		}
		file.Write(chunk)
	}
	return uint32(file.Len()),
		ioutil.WriteFile(
			downloads+filename,
			file.Bytes(),
			os.ModePerm,
		)
}
