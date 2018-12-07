package chain

import (
	"fmt"
	"errors"
	"bytes"
	"strings"
)

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

func (block BlockPublish) DecrementHopLimit() {
	block.HopLimit--
}

func (block BlockPublish) GetHopLimit() uint32 {
	return block.HopLimit
}

type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

const blockHopLimit = 20

var BlocksChan = make(chan BlockPublish)

func ReceiveBlock(block Block) {
	if !block.verifyHash() {
		fmt.Println("received malicious block with hash", block.Hash())
		return
	}
	if hasBlock(&block) {
		fmt.Println("already has block", block.Hash())
	}
	if !hasParentOf(&block) && getHeadsCount() > 1 {
		fmt.Println("received block with no parents in chain", block.Hash())
		return
	}
	var head *[32]byte = nil
	if isLongest(block.PrevHash) {
		head = &block.PrevHash
	}
	if !block.canAddBlockToHead(head) {
		fmt.Println("cannot add block", block.Hash())
	}

	pauseMining()
	err := addBlock(block)
	if err != nil {
		fmt.Println("error adding block", block.Hash(), err)
	}
	removeClaimedPendingTransactions()
	go Mine()
}

func (block *Block) canAddBlockToLedger() bool {
	for _, tx := range block.Transactions {
		if isNameClaimed(tx.File.Name) {
			return false
		}
	}
	return true
}

func (block *Block) canAddBlockToHead(head *[32]byte) bool {
	if head == nil {
		return block.canAddBlockToLedger()
	}
	txs := make(map[string]struct{}, len(block.Transactions))
	for _, tx := range block.Transactions {
		txs[tx.File.Name] = struct{}{}
	}
	canAdd, err := canAddFilenamesToHead(txs, *head)
	if err != nil {
		fmt.Printf(
			"cannot add block %s to head %s: %s",
			block.Hash(),
			*head,
			err.Error(),
		)
	}
	return canAdd
}

func canAddFilenamesToHead(
	txs map[string]struct{},
	head [32]byte,
) (bool, error) {
	if head == genesis {
		return true, nil
	}
	blockchain.RLock()
	headBlock, hasBlock := blockchain.m[head]
	if !hasBlock {
		return true, nil
	}
	blockchain.RUnlock()
	if isConflictingBlock(txs, &headBlock) {
		return false, errors.New("conflicting block")
	}
	return canAddFilenamesToHead(txs, headBlock.PrevHash)
}

func isConflictingBlock(txs map[string]struct{}, other *Block) bool {
	for _, tx := range other.Transactions {
		if _, hasTx := txs[tx.File.Name]; hasTx {
			return true
		}
	}
	return false
}

func (block *Block) verifyHash() bool {
	hash := block.Hash()
	return bytes.Equal(hash[:2], zeroHash)
}

func hasParentOf(block *Block) bool {
	blockchain.RLock()
	defer blockchain.RUnlock()
	_, hasParent := blockchain.m[block.PrevHash]
	return hasParent
}

func publishBlock(block Block) {
	BlocksChan <- BlockPublish{
		Block:    block,
		HopLimit: blockHopLimit,
	}
}

func (block *Block) toString() string {
	return fmt.Sprintf(
		"%x:%x:%s",
		block.Hash(),
		block.PrevHash,
		block.describeTransactions(),
	)
}

func describeBlock(hash [32]byte) string {
	blockchain.RLock()
	defer blockchain.RUnlock()
	block, hasBlock := blockchain.m[hash]
	if hash == genesis || !hasBlock {
		return fmt.Sprintf("%x", hash)
	}
	return fmt.Sprintf(
		"%x:%x:%s",
		hash,
		block.PrevHash,
		block.describeTransactions(),
	)
}

func (block *Block) describeTransactions() string {
	filenames := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		filenames[i] = tx.File.Name
	}
	return strings.Join(filenames, ",")
}