package message

type Operation struct {
	Rumor      *RumorRequest
	Identifier *IdentifierRequest
}

type RumorRequest struct {
	Contents string
}

type IdentifierRequest struct{}

type IdentifierResponse struct {
	Identifier string
}

type PeersRequest struct{}

type PeersResponse struct {
	Peers []string
}

type MessagesRequest struct {
	StartIndex int
}

type PeerMessages struct {
	Peer     string
	Messages []RumorMessage
}

type MessagesResponse struct {
	StartIndex int
	Messages   []RumorMessage
}

type ValidationResponse struct {
	Success bool
}

type AddPeerRequest struct {
	Address string
	Port    string
}

type PrivatePutRequest struct {
	Contents    string
	Destination string
}

type PrivateGetRequest struct {
	Partner        string
	UnorderedIndex int
	OrderedIndex   int
}

type PrivateGetResponse struct {
	Unordered      []PrivateMessage
	Ordered        []PrivateMessage
	UnorderedIndex int
	OrderedIndex   int
}

type ChatsRequest struct{}

type ChatsResponse struct {
	Origins []string
}

type FileShareRequest struct {
	Name     string
	Size     uint32
	Metafile []byte
	Metahash []byte
}
