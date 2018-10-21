package message

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type GossipPacket struct {
	Simple  *SimpleMessage
	Rumor   *RumorMessage
	Status  *StatusPacket
	Private *PrivateMessage
}

type ClientPacket struct {
	Rumor       *RumorRequest
	Identifier  *IdentifierRequest
	Peers       *PeersRequest
	Messages    *MessagesRequest
	AddPeer     *AddPeerRequest
	SendPrivate *PrivatePutRequest
	GetPrivate  *PrivateGetRequest
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}
