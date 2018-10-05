package gossip

import (
	"sync"
	"net"
)

type Peers struct {
	sync.RWMutex
	m map[string]*net.UDPAddr
}