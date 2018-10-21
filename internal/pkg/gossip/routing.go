package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"net"
	"fmt"
	"time"
)

const hopLimit = 10

type RouteInfo struct {
	nextHop        *net.UDPAddr
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
			nextHop: sender, sequenceNumber: rumor.ID,
		}
		fmt.Printf("DSDV %s %s\n", rumor.Origin, sender.String())
	}
	gossiper.routing.Unlock()
}

func (gossiper *Gossiper) routeRumorMessages(
	rtimer int,
) {
	if rtimer == 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(rtimer) * time.Second)
	defer ticker.Stop()
	peer := gossiper.pickRumormongeringPartner(map[string]struct{}{})
	gossiper.sendRouteRumor(peer)
	for {
		_ = <-ticker.C
		peer := gossiper.pickRumormongeringPartner(map[string]struct{}{})
		gossiper.sendRouteRumor(peer)
	}
}

func (gossiper *Gossiper) sendRouteRumor(peer *net.UDPAddr) {
	gossiper.routing.Lock()
	sequenceNumber := gossiper.routing.m[gossiper.Name].sequenceNumber + 1
	gossiper.routing.m[gossiper.Name] = RouteInfo{
		nextHop:        gossiper.gossipAddr,
		sequenceNumber: sequenceNumber,
	}
	gossiper.routing.Unlock()
	emptyRumor := &message.RumorMessage{
		Origin: gossiper.Name,
		ID:     sequenceNumber,
		Text:   "",
	}
	gossiper.memorizeRumor(emptyRumor)
	bytes := encodeMessage(&message.GossipPacket{Rumor: emptyRumor})
	gossiper.gossipConn.WriteToUDP(bytes, peer)
}

func isRouteRumor(rumor *message.RumorMessage) bool {
	return len(rumor.Text) == 0
}