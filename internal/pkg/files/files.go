package files

import (
	"encoding/base64"
	"crypto/sha256"
)

type File struct {
	Name     string
	Size     uint32
	Metafile []byte
	Metahash []byte
}

func KeyToFilename(key []byte) string {
	return base64.URLEncoding.EncodeToString(key)
}

func HashChunk(chunk []byte) []byte {
	hasher := sha256.New()
	hasher.Write(chunk)
	return hasher.Sum(nil)
}
