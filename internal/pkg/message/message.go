package message

import "github.com/aounleonardo/Peerster/internal/pkg/chain"

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string

	RaspRequest *chain.RaspRequest
}

type GossipPacket struct {
	Simple        *SimpleMessage
	Rumor         *RumorMessage
	Status        *StatusPacket
	Private       *PrivateMessage
	DataRequest   *DataRequest
	DataReply     *DataReply
	SearchRequest *SearchRequest
	SearchReply   *SearchReply
	TxPublish     *chain.TxPublish
	BlockPublish  *chain.BlockPublish
}

type ClientPacket struct {
	Rumor       *RumorRequest
	Identifier  *IdentifierRequest
	Peers       *PeersRequest
	Messages    *MessagesRequest
	AddPeer     *AddPeerRequest
	SendPrivate *PrivatePutRequest
	GetPrivate  *PrivateGetRequest
	Chats       *ChatsRequest
	FileShare   *FileShareRequest
	Download    *FileDownloadRequest
	Search      *PerformSearchRequest
	GetSearches *SearchesRequest

	CreateMatch *chain.CreateMatchRequest
	AcceptMatch *chain.AcceptMatchRequest

	TestReconstruct *TestFileReconstructRequest
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

	Rasp *RaspMessage
}

type RaspMessage struct {
	Request  *chain.RaspRequest
	Response *chain.RaspResponse
	Attack   *chain.RaspAttack
	Defence  *chain.RaspDefence
}

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}
