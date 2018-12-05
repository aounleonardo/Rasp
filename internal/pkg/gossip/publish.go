package gossip

import "github.com/aounleonardo/Peerster/internal/pkg/chain"

func publishBlocks() {
	for {
		select {
		case block := <-chain.BlocksChan:
			advertiseBlock(block)
		}
	}
}

func advertiseBlock(block chain.BlockPublish) {

}
