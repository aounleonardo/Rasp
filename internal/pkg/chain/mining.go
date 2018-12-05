package chain

import (
	"fmt"
	"math/rand"
)

var stopMining = make(chan struct{})
var zeroHash = ([16]byte{})[:]

func Mine() {
	for hasNoPendingTransactions() {
		select {
		case <-stopMining:
			return
		default:
		}
	}
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
		case <-stopMining:
			fmt.Println("is stopped")
			return
		default:
			newBlock.Nonce = getRandomNonce()
			if newBlock.verifyHash() {
				ReceiveBlock(newBlock)
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
		fmt.Println("sent stopMining")
	default:
		fmt.Println("tried to stop but already stopped")
	}

}