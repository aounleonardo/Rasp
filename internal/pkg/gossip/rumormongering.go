package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"time"
	"math/rand"
	"fmt"
	"net"
)

func (gossiper *Gossiper) receiveRumorPacket(
	rumor *message.RumorMessage,
	sender *net.UDPAddr,
) {
	fmt.Printf(
		"RUMOR origin %s from %s ID %d contents %s\n",
		rumor.Origin,
		sender.String(),
		rumor.ID,
		rumor.Text,
	)

	if rumor.ID == gossiper.nextIdForPeer(rumor.Origin) {
		gossiper.wants.Lock()
		gossiper.rumors.Lock()
		gossiper.rumors.m[rumor.Origin][rumor.ID] = rumor
		gossiper.wants.m[rumor.Origin] = rumor.ID + 1
		gossiper.wants.Unlock()
		gossiper.rumors.Unlock()
		go gossiper.rumormonger(rumor, sender)
	}

	gossiper.sendStatusPacket(sender)
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
	var filteredPeers []string
	gossiper.peers.RLock()
	defer gossiper.peers.RUnlock()
	for peer := range gossiper.peers.m {
		_, shouldFilter := except[peer]
		if !shouldFilter {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	if len(filteredPeers) == 0 {
		return nil
	}

	n := rand.Intn(len(filteredPeers))
	return gossiper.peers.m[filteredPeers[n]]
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
	fmt.Printf("MONGERING with %s\n", peer.String())
	gossiper.gossipConn.WriteToUDP(bytes, peer)
	expectedAcks[peer.String()] ++

	for {
		var operation int
		var missing message.PeerStatus
		timer := time.NewTimer(time.Second)
		select {
		case ack := <-acks[peer.String()]:
			operation, missing = gossiper.compareStatuses(ack)
		case <-timer.C:
			operation, missing = NOP, message.PeerStatus{}
			if expectedAcks[peer.String()] > 0 {
				expectedAcks[peer.String()] --
			}
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
					fmt.Printf(
						"FLIPPED COIN sending rumor to %s\n",
						newPartner.String(),
					)
					gossiper.rumormongerWith(rumor, newPartner, sender)
				}
			}
			return
		}
	}
}
