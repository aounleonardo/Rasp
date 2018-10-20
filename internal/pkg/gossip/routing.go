package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"net"
	"fmt"
	"time"
)

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
	ticker := time.NewTicker(
		time.Duration(float64(rtimer) * time.Second.Seconds()),
	)
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
	gossiper.wants.Lock()
	gossiper.wants.m[gossiper.Name] += 1
	gossiper.wants.Unlock()

	gossiper.routing.Lock()
	sequenceNumber := gossiper.routing.m[gossiper.Name].sequenceNumber + 1
	gossiper.routing.m[gossiper.Name] = RouteInfo{
		nextHop:        gossiper.gossipAddr,
		sequenceNumber: sequenceNumber,
	}
	gossiper.routing.Unlock()
	packet := &message.GossipPacket{
		Rumor: &message.RumorMessage{
			Origin: gossiper.Name,
			ID:     sequenceNumber,
			Text:   "",
		},
	}
	bytes := encodeMessage(packet)
	gossiper.gossipConn.WriteToUDP(bytes, peer)
}
