package chain

import (
	"fmt"
	"errors"
)

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

const blockHopLimit = 20

var BlocksChan = make(chan BlockPublish)

func ReceiveBlock(block BlockPublish) {

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
) (bool, error){
	if head == genesis {
		return true, nil
	}
	blockchain.RLock()
	headBlock, hasBlock := blockchain.m[head]
	if !hasBlock {
		return false, errors.New("missing block in chain")
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

func publishBlock(block Block) {

}
