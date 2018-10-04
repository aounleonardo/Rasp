package message

type Operation struct {
	Rumor *RumorRequest
	Identifier *IdentifierRequest
}

type RumorRequest struct {
	Contents string
}

type IdentifierRequest struct {}

type IdentifierResponse struct {
	Identifier string
}

type PeersRequest struct {}

type PeersResponse struct {
	Peers []string
}

type MessagesRequest struct {
	Status StatusPacket
}

type PeerMessages struct {
	Peer string
	Messages []RumorMessage
}

type MessagesResponse struct {
	Messages []PeerMessages
}