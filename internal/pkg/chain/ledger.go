package chain

import "sync"

var ledger = struct {
	sync.RWMutex
	m map[string]File
}{
	m: make(map[string]File),
}

var blockchain = struct {
	sync.RWMutex
	m map[[32]byte]Block
}{
	m: make(map[[32]byte]Block),
}

func isNameClaimed(filename string) bool {
	ledger.RLock()
	defer ledger.RUnlock()
	_, hasTx := ledger.m[filename]
	return hasTx
}