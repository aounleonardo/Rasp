package chain

import (
	"fmt"
	"math/rand"
)

var stop = make(chan struct{})
var zeroHash = ([16]byte{})[:]

func Mine() {
	txs := getNewTransactions()
	blockchain.RLock()
	newBlock := Block{
		PrevHash: blockchain.longest,
		Nonce: [32]byte{},
		Transactions: txs,
	}
	blockchain.RUnlock()
	for true {
		select {
		case <-stop:
			fmt.Println("is stopped")
			return
		default:
			newBlock.Nonce = getRandomNonce()
			if newBlock.verifyHash() {

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
