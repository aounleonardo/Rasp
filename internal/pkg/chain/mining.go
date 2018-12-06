package chain

import (
	"fmt"
	"math/rand"
	"time"
)

var stopMining = make(chan struct{}, 1)
var zeroHash = make([]byte, 2)

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

	miningStartTime := time.Now()
	for {
		select {
		case <-stopMining:
			return
		default:
			newBlock.Nonce = getRandomNonce()
			if newBlock.verifyHash() {
				fmt.Printf("FOUND-BLOCK [%x]\n", newBlock.Hash())
				ReceiveBlock(newBlock)
				miningDuration := time.Now().Sub(miningStartTime)
				go func() {
					time.Sleep(2 * miningDuration)
					publishBlock(newBlock)
				}()
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