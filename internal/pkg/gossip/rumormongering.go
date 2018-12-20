package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"math/rand"
	"net"
	"time"
)

func (gossiper *Gossiper) receiveRumorPacket(
	rumor *message.RumorMessage,
	sender *net.UDPAddr,
) {
	gossiper.updateNextHop(rumor, sender)
	if rumor.ID == gossiper.nextIdForPeer(rumor.Origin) {
		gossiper.memorizeRumor(rumor)
		go gossiper.rumormonger(rumor, sender)
	}
	gossiper.sendStatusPacket(sender)

	if rumor.RaspRequest == nil {
		return
	}

}

func (gossiper *Gossiper) rumormonger(
	rumor *message.RumorMessage,
	sender *net.UDPAddr,
) {
	selectedPeerAddr := gossiper.pickRumormongeringPartner(
		map[string]struct{}{sender.String(): {}},
	)
	if selectedPeerAddr == nil {
		return
	}
	gossiper.rumormongerWith(rumor, selectedPeerAddr, sender)
}

func (gossiper *Gossiper) pickRumormongeringPartner(
	except map[string]struct{},
) *net.UDPAddr {
	filteredPeers := gossiper.getFilteredPeers(except)

	if len(filteredPeers) == 0 {
		return nil
	}

	n := rand.Intn(len(filteredPeers))
	return gossiper.peers.m[filteredPeers[n]]
}

func (gossiper *Gossiper) getFilteredPeers(
	except map[string]struct{},
) []string {
	var filteredPeers []string
	gossiper.peers.RLock()
	defer gossiper.peers.RUnlock()
	for peer := range gossiper.peers.m {
		_, shouldFilter := except[peer]
		if !shouldFilter {
			filteredPeers = append(filteredPeers, peer)
		}
	}
	return filteredPeers
}

func (gossiper *Gossiper) rumormongerWith(
	rumor *message.RumorMessage,
	peer *net.UDPAddr,
	sender *net.UDPAddr,
) {
	bytes := encodeMessage(&message.GossipPacket{Rumor: rumor})
	if bytes == nil {
		return
	}
	gossiper.gossipConn.WriteToUDP(bytes, peer)
	acks.Lock()
	acks.expected[peer.String()]++
	acks.Unlock()

	for {
		var operation int
		var missing message.PeerStatus
		timer := time.NewTimer(time.Second)
		select {
		// TODO should I lock here?
		case ack := <-acks.queue[peer.String()]:
			operation, missing = gossiper.compareStatuses(ack)
		case <-timer.C:
			operation, missing = NOP, message.PeerStatus{}
			acks.Lock()
			if acks.expected[peer.String()] > 0 {
				acks.expected[peer.String()]--
			}
			acks.Unlock()
		}
		switch operation {
		case SEND:
			gossiper.sendMissingRumor(&missing, peer)
		case REQUEST:
			gossiper.sendStatusPacket(peer)
			return
		case NOP:
			if rand.Intn(2) == 0 {
				newPartner := gossiper.pickRumormongeringPartner(
					map[string]struct{}{peer.String(): {}},
				)
				if newPartner != nil {
					gossiper.rumormongerWith(rumor, newPartner, sender)
				}
			}
			return
		}
	}
}
