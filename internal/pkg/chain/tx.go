package chain

import (
	"sync"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"errors"
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

const txHopLimit = 10

var pendingTransactions = struct {
	sync.RWMutex
	l []TxPublish
}{
	l: make([]TxPublish, 0),
}

func (tx *File) shouldDiscardTransaction() bool {
	if isNameClaimed(tx.Name) {
		return true
	}

	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()
	for _, pending := range pendingTransactions.l {
		if tx.Name == pending.File.Name {
			return true
		}
	}
	return false
}

func ReceiveTransaction(tx TxPublish) {
	if tx.File.shouldDiscardTransaction() {
		return
	}

	pendingTransactions.Lock()
	pendingTransactions.l = append(pendingTransactions.l, tx)
	pendingTransactions.Unlock()
}

func BuildTransaction(
	filename string,
	filesize uint64,
	metakey string,
) (TxPublish, error) {
	size := int64(filesize)
	if size < 0 {
		return TxPublish{}, errors.New("invalid file size")
	}
	file := File{
		Name:         filename,
		Size:         size,
		MetafileHash: files.KeyToHash(metakey),
	}
	tx := TxPublish{File: file, HopLimit: txHopLimit}
	go ReceiveTransaction(tx)
	return tx, nil
}

func removeClaimedPendingTransactions() {
	pendingTransactions.Lock()
	ledger.RLock()

	newPendings := make([]TxPublish, 0)
	for _, tx := range pendingTransactions.l {
		if !isNameClaimed(tx.File.Name) {
			newPendings = append(newPendings, tx)
		}
	}
	pendingTransactions.l = newPendings

	ledger.RUnlock()
	pendingTransactions.Unlock()
}
