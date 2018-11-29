package chain

import (
	"sync"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
)

type TxPublish struct {
	File     File
	HopLimit uint32
}

type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

var pendingTransactions = struct {
	sync.RWMutex
	l []File
}{
	l: make([]File, 0),
}

func shouldDiscardTransaction(tx *File) bool {
	if isNameClaimed(tx.Name) {
		return true
	}

	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()
	for _, pending := range pendingTransactions.l {
		if tx.Name == pending.Name {
			return true
		}
	}
	return false
}

func ReceiveTransaction(tx File) {
	if shouldDiscardTransaction(&tx) {
		return
	}

	pendingTransactions.Lock()
	pendingTransactions.l = append(pendingTransactions.l, tx)
	pendingTransactions.Unlock()
}

func BuildTransaction(filename string, filesize uint64, metakey string)  {
	size := int64(filesize)
	if size < 0 {
		return
	}
	tx := File{
		Name: filename,
		Size: size,
		MetafileHash: files.KeyToHash(metakey),
	}
	ReceiveTransaction(tx)
}