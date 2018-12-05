package chain

import (
	"sync"
)

type TxPublish struct {
	File     File
	HopLimit uint32
}

func (tx TxPublish) DecrementHopLimit() {
	tx.HopLimit--
}

func (tx TxPublish) GetHopLimit() uint32 {
	return tx.HopLimit
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

func hasNoPendingTransactions() bool {
	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()

	return len(pendingTransactions.l) < 1
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
	file File,
) TxPublish {
	tx := TxPublish{File: file, HopLimit: txHopLimit}
	go ReceiveTransaction(tx)
	return tx
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
