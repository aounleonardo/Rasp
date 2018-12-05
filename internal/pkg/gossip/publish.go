package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/chain"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"fmt"
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
)

func (gossiper *Gossiper) publishMinedBlocks() {
	for {
		select {
		case block := <-chain.BlocksChan:
			gossiper.advertisePublisher(block, nil)
		}
	}
}

type Publish interface {
	DecrementHopLimit()
	GetHopLimit() uint32
}

func getGossipPacket(pub Publish) *message.GossipPacket {
	switch t := pub.(type) {
	case chain.BlockPublish:
		return &message.GossipPacket{BlockPublish: &t}
	case chain.TxPublish:
		return &message.GossipPacket{TxPublish: &t}
	default:
		fmt.Println("unknown Publish type", pub)
		return nil
	}
}

func (gossiper *Gossiper) advertisePublisher(
	pub Publish,
	sender *string,
) {
	if pub.GetHopLimit() == 0 {
		return
	}
	pub.DecrementHopLimit()

	except := make(map[string]struct{})
	if sender != nil {
		except[*sender] = struct{}{}
	}
	filteredPeers := gossiper.getFilteredPeers(except)
	for _, peer := range filteredPeers {
		bytes := encodeMessage(getGossipPacket(pub))
		gossiper.peers.RLock()
		gossiper.gossipConn.WriteToUDP(bytes, gossiper.peers.m[peer])
		gossiper.peers.RUnlock()
	}
}

func (gossiper *Gossiper) receiveTxPublish(
	tx *chain.TxPublish,
	fromSender *net.UDPAddr,
) {
	chain.ReceiveTransaction(*tx)
	var sender *string
	*sender = fromSender.String()
	gossiper.advertisePublisher(Publish(*tx), sender)
}

func (gossiper *Gossiper) indexFile(file *files.File) {
	chain.ReceiveTransaction(chain.BuildTransaction(chain.File{
		Name:         file.Name,
		Size:         int64(file.Size),
		MetafileHash: file.Metahash,
	}))
}
