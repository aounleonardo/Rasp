package chain

import "sync"

var ledger = struct {
	sync.RWMutex
	m map[string]File
}{
	m: make(map[string]File),
}

var genesis = [32]byte{}

var blockchain = struct {
	sync.RWMutex
	m map[[32]byte]Block
	heads [][32]byte
	longest [32]byte
}{
	m: make(map[[32]byte]Block),
	heads: make([][32]byte, 0),
	longest: genesis,
}

func isNameClaimed(filename string) bool {
	ledger.RLock()
	defer ledger.RUnlock()
	_, hasTx := ledger.m[filename]
	return hasTx
}