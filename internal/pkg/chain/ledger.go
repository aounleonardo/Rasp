package chain

import (
	"sync"
	"errors"
	"fmt"
)

var ledger = struct {
	sync.RWMutex
	m map[string]File
}{
	m: make(map[string]File),
}

var genesis = [32]byte{}

var blockchain = struct {
	sync.RWMutex
	m       map[[32]byte]Block
	heads   map[[32]byte]struct{}
	longest [32]byte
	length  int
}{
	m:       map[[32]byte]Block{genesis: {}},
	heads:   map[[32]byte]struct{}{genesis: {}},
	longest: genesis,
	length:  0,
}

func isNameClaimed(filename string) bool {
	ledger.RLock()
	defer ledger.RUnlock()
	_, hasTx := ledger.m[filename]
	return hasTx
}

func getHeadsCount() int {
	blockchain.RLock()
	defer blockchain.RUnlock()
	return len(blockchain.heads)
}

func isLongest(head [32]byte) bool {
	blockchain.RLock()
	defer blockchain.RUnlock()
	return head == blockchain.longest
}

func hasBlock(block *Block) bool {
	blockchain.RLock()
	defer blockchain.RUnlock()
	_, hasBlock := blockchain.m[block.Hash()]
	return hasBlock
}

func getBlock(hash [32]byte) (*Block, error) {
	blockchain.RLock()
	defer blockchain.RUnlock()

	if block, hasBlock := blockchain.m[hash]; hasBlock {
		return &block, nil
	}
	return nil, errors.New(fmt.Sprintf("does not have block %s", hash))
}

func addBlock(block Block) error {
	upsertHead(block.PrevHash)
	hash := block.Hash()

	blockchain.Lock()
	blockchain.m[hash] = block
	delete(blockchain.heads, block.PrevHash)
	blockchain.heads[hash] = struct{}{}
	blockchain.Unlock()

	err := switchHeadTo(hash)
	if err != nil {
		return err
	}
	return nil
}

func switchHeadTo(hash [32]byte) error {
	currentHead, currentLength := getCurrentHead()
	newLength := getHeadLength(hash)

	if currentLength >= newLength {
		return nil
	}

	ancestor := getCommonAncestor(currentHead, hash)
	if ancestor != currentHead {
		stop, err := rollbackTo(ancestor)
		if err != nil {
			_, fallback := applyChangesUpTo(currentHead, stop)
			if fallback != nil {
				errors.New(fmt.Sprintf(
					"got error %s while applying changes,"+
						" tried to fallback but %s",
					err.Error(),
					fallback.Error(),
				))
			}
			return err
		}
	}

	_, err := applyChangesUpTo(hash, ancestor)
	if err != nil {
		rollbackTo(ancestor)
		_, fallback := applyChangesUpTo(currentHead, ancestor)
		if fallback != nil {
			return errors.New(fmt.Sprintf(
				"got error %s while applying changes,"+
					" tried to fallback but %s",
				err.Error(),
				fallback.Error(),
			))
		}
		return err
	}
	blockchain.Lock()
	blockchain.longest = hash
	blockchain.length = newLength
	blockchain.Unlock()
	return nil
}

func getCommonAncestor(block [32]byte, other [32]byte) [32]byte {
	blockchain.RLock()
	defer blockchain.RUnlock()

	pathToRoot := getChainHashes(block)
	ancestor, err := findFirstInPath(other, pathToRoot)
	if err != nil {
		fmt.Println("error when searching for common ancestor", block, other)
		return block
	}
	return ancestor
}

func findFirstInPath(start [32]byte, path [][32]byte) ([32]byte, error) {
	nodesInPath := make(map[[32]byte]struct{}, len(path))
	for _, node := range path {
		nodesInPath[node] = struct{}{}
	}
	blockchain.RLock()
	defer blockchain.RUnlock()

	node := start
	for {
		block, hasNode := blockchain.m[node]
		if !hasNode {
			return [32]byte{},
				errors.New("unexpected error: can't find node")
		}
		if _, inPath := nodesInPath[node]; inPath {
			return node, nil
		}
		node = block.PrevHash
	}
	return [32]byte{},
		errors.New("unexpected error: can't find node")
}

func getCurrentHead() ([32]byte, int) {
	blockchain.RLock()
	defer blockchain.RUnlock()

	return blockchain.longest, blockchain.length
}

func upsertHead(hash [32]byte) {
	blockchain.Lock()
	defer blockchain.Unlock()

	if _, hasHead := blockchain.heads[hash]; hasHead {
		return
	}
	blockchain.heads[hash] = struct{}{}
}

func getHeadLength(hash [32]byte) int {
	return len(getChainHashes(hash))
}

func getChainHashes(start [32]byte) [][32]byte {
	chain := make([][32]byte, 0)
	blockchain.RLock()
	defer blockchain.RUnlock()

	node := start
	for {
		block, hasNode := blockchain.m[node]
		if !hasNode {
			return chain
		}
		chain = append(chain, node)
		if node == genesis {
			return chain
		}
		node = block.PrevHash
	}
	return chain
}

func rollbackTo(hash [32]byte) ([32]byte, error) {
	currentHead, _ := getCurrentHead()
	pathToRoot := getChainHashes(currentHead)
	index := 0
	node := pathToRoot[0]
	for ; index < len(pathToRoot) && node != hash;
	index, node = index+1, pathToRoot[index+1] {
		err := denyBlock(node)
		if err != nil {
			return node, err
		}
	}
	fmt.Printf("FORK-LONGER rewind %d blocks", index)
	return node, nil
}

func denyBlock(hash [32]byte) error {
	block, err := getBlock(hash)
	if err != nil {
		return err
	}
	ledger.Lock()
	pendingTransactions.Lock()
	for _, tx := range block.Transactions {
		delete(ledger.m, tx.File.Name)
		pendingTransactions.l = append(pendingTransactions.l, tx)
	}
	pendingTransactions.Unlock()
	ledger.Unlock()
	return nil
}

func applyChangesUpTo(stop [32]byte, ancestor [32]byte) ([32]byte, error) {
	pathToRoot := getChainHashes(stop)
	if len(pathToRoot) < 1 {
		return genesis, errors.New(
			fmt.Sprintf("cannot find path to root from %s", stop),
		)
	}
	ancestorIndex, err := findNodeInPath(pathToRoot, ancestor)
	if err != nil {
		return genesis, err
	}
	changesToApply := pathToRoot[:ancestorIndex]
	for _, node := range changesToApply {
		err := applyBlock(node)
		if err != nil {
			return node, err
		}
	}

	return stop, nil
}

func applyBlock(hash [32]byte) error {
	block, err := getBlock(hash)
	if err != nil {
		return err
	}
	ledger.Lock()
	for _, tx := range block.Transactions {
		if _, nameIsClaimed := ledger.m[tx.File.Name]; !nameIsClaimed {
			ledger.m[tx.File.Name] = tx.File
		}
	}
	ledger.Unlock()
	return nil
}

func findNodeInPath(path [][32]byte, node [32]byte) (int, error) {
	for searchIndex := 0; searchIndex < len(path); searchIndex++ {
		if path[searchIndex] == node {
			return searchIndex, nil
		}
	}
	return -1, errors.New(fmt.Sprintf(
		"cannot find node %x in path %x",
		node,
		path,
	))
}
