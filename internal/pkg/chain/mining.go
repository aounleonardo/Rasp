package chain

import (
	"fmt"
	"math/rand"
)

var stopMining = make(chan struct{}, 1)
var zeroHash = make([]byte, 2)

func Mine() {
	for hasNoPendingTransactions() {
		select {
		case <-stopMining:
			return
		}
	}
	txs := getNewTransactions()
	blockchain.RLock()
	newBlock := Block{
		PrevHash:     blockchain.longest,
		Nonce:        [32]byte{},
		Transactions: txs,
	}
	blockchain.RUnlock()
	for {
		select {
		case <-stopMining:
			return
		default:
			newBlock.Nonce = getRandomNonce()
			if newBlock.verifyHash() {
				fmt.Println("FOUND-BLOCK", newBlock.Hash())
				ReceiveBlock(newBlock)
				publishBlock(newBlock)
			}
		}
	}
}

func getNewTransactions() []TxPublish {
	pendingTransactions.RLock()
	newTransactions := pendingTransactions.l
	pendingTransactions.RUnlock()
	return newTransactions
}

func getRandomNonce() [32]byte {
	nonce := make([]byte, 32)
	rand.Read(nonce)
	var ret [32]byte
	copy(ret[:], nonce)
	return ret
}

func pauseMining() {
	select {
	case stopMining <- struct{}{}:
	default:
	}
}