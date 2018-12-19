package chain

import (
	"fmt"
	"math/rand"
	"time"
)

var stopMining = make(chan struct{}, 1)
var zeroHash = make([]byte, 2)
var first = true

func Mine() {
	for hasNoPendingTransactions() {
		select {
		case <-stopMining:
			return
		default:
			time.Sleep(time.Second)
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
				fmt.Printf("FOUND-BLOCK %x\n", newBlock.Hash())
				ReceiveBlock(newBlock)
				go publishBlock(newBlock)
			}
		}
	}
}

func getNewTransactions() []TxPublish {
	var newTransactions = []TxPublish(nil)
	var tmpBalances = make(map[string]int64)
	spawns := getNewSpawns(tmpBalances)
	newTransactions = append(newTransactions, spawns...)
	attacks := getNewAttacks(tmpBalances)
	newTransactions = append(newTransactions, attacks...)
	defences := getNewDefences(attacks)
	newTransactions = append(newTransactions, defences...)
	reveals := getNewReveals(defences)
	newTransactions = append(newTransactions, reveals...)
	cancels := getNewCancels(defences)
	newTransactions = append(newTransactions, cancels...)
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
