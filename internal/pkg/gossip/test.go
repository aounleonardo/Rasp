package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"net"
)

func (gossiper *Gossiper) handleTestPacket(
	packet *message.ClientPacket,
	clientAddr *net.UDPAddr,
) {
	if packet.TestReconstruct != nil {
		gossiper.handleTestReconstruct(packet.TestReconstruct, clientAddr)
	}
}

func (gossiper *Gossiper) handleTestReconstruct(
	request *message.TestFileReconstructRequest,
	clientAddr *net.UDPAddr,
) {
	hashEncoding := files.HashToKey(request.Metahash)
	err := files.TestCombineChunksIntoFile(hashEncoding, request.Filename)
	gossiper.sendToClient(
		&message.ValidationResponse{Success: err != nil},
		clientAddr,
	)
}
