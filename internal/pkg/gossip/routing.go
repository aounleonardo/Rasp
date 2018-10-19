package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"net"
	"fmt"
)

type RouteInfo struct {
	nextHop        string
	sequenceNumber uint32
}

func (gossiper *Gossiper) updateNextHop(
	rumor *message.RumorMessage,
	sender *net.UDPAddr,
) {
	gossiper.routing.Lock()
	if route, hasRoute := gossiper.routing.m[rumor.Origin];
		!hasRoute ||
			rumor.ID > route.sequenceNumber {
		gossiper.routing.m[rumor.Origin] = RouteInfo{
			nextHop: sender.String(), sequenceNumber: rumor.ID,
		}
		fmt.Printf("DSDV %s %s\n", rumor.Origin, sender.String())
	}
	gossiper.routing.Unlock()
}
