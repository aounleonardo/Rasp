package chain

import (
	"bytes"
	"fmt"
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
		fmt.Println("received block with invalid POW hash", block.Hash())
		return
	}
	if hasBlock(&block) {
		fmt.Println("already has block", block.Hash())
		return
	}
	if !hasParent(&block) {
		fmt.Printf(
			"parent %x of block %x does not exist\n",
			block.PrevHash,
			block.Hash(),
		)
		return

	}

	head := block.PrevHash
	if !block.canAddBlockToUpsertedHead(head) {
		fmt.Println("cannot add block", block.Hash())
	}

	pauseMining()
	blockchain.Lock()
	addBlockUnsafe(block)
	removeClaimedPendingTransactionsUnsafe()
	blockchain.Unlock()
	go Mine()
}

// Unsafe lock blockchain before using
func (block *Block) canAddBlockToLedgerUnsafe(ledger ledger) bool {
	var currentStage = Spawn
	var tmpBalances = getBalancesUnsafe(ledger)
	var attacks = make(map[Uid]struct{})
	var defences = make(map[Uid]struct{})
	var reveals = make(map[Uid]struct{})
	var cancels = make(map[Uid]struct{})
	for _, tx := range block.Transactions {
		if tx.Action.Type < currentStage {
			return false
		}
		currentStage = tx.Action.Type
		if canAdd := tx.canAddToLedgerUnsafe(
			ledger,
			tmpBalances,
			attacks,
			defences,
			reveals,
			cancels,
		); !canAdd {
			return false
		}
	}
	return true
}

func (block *Block) canAddBlockToUpsertedHead(head [32]byte) bool {
	blockchain.Lock()
	defer blockchain.Unlock()
	if ledger, exists := blockchain.heads[head]; exists {
		return block.canAddBlockToLedgerUnsafe(ledger)
	}
	var forkTxs = map[Stage]map[Uid]GameAction{
		Spawn:   make(map[Uid]GameAction),
		Attack:  make(map[Uid]GameAction),
		Defence: make(map[Uid]GameAction),
		Reveal:  make(map[Uid]GameAction),
		Cancel:  make(map[Uid]GameAction),
	}
	newLedger := createForkLedgerUnsafe(forkTxs, head, 0)
	blockchain.heads[head] = newLedger
	return block.canAddBlockToLedgerUnsafe(newLedger)
}

func (block *Block) verifyHash() bool {
	hash := block.Hash()
	return bytes.Equal(hash[:2], zeroHash)
}

func publishBlock(block Block) {
	BlocksChan <- BlockPublish{
		Block:    block,
		HopLimit: blockHopLimit,
	}
}
