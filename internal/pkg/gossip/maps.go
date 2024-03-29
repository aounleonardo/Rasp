package gossip

import (
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"net"
	"sync"
)

type Peers struct {
	sync.RWMutex
	m map[string]*net.UDPAddr
}

type Needs struct {
	sync.RWMutex
	m map[string]uint32
}

type Rumors struct {
	sync.RWMutex
	m map[string]map[uint32]*message.RumorMessage
}

type Acks struct {
	sync.RWMutex
	queue    map[string]chan *message.StatusPacket
	expected map[string]int
}

type Routes struct {
	sync.RWMutex
	m map[string]RouteInfo
}

type Privates struct {
	sync.RWMutex
	m map[string]*ChatHistory
}

type Files struct {
	sync.RWMutex
	m map[string]files.File
}
