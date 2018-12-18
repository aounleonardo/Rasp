package message

import "github.com/aounleonardo/Peerster/internal/pkg/chain"

type Signature = []byte

type RaspRequest struct {
	Identifier  chain.Uid
	Bet         chain.Bet
	Destination *string
	Origin      string
	Signature   Signature
}

type RaspResponse struct {
	Destination string
	Origin      string
	Identifier  chain.Uid
	Signature   Signature
}

type RaspAttack struct {
	Destination string
	Origin      string
	Identifier  chain.Uid
	Bet         chain.Bet
	SignedBet   Signature
	SignedMove  Signature
}

type RaspDefence struct {
	Destination string
	Origin      string
	Identifier  chain.Uid
	Move        chain.Move
	SignedMove  Signature
}

type CreateMatchRequest struct {
	Destination *string
	Bet         chain.Bet
	Move        chain.Move
}

type AcceptMatchRequest struct {
	Identifier chain.Uid
	Move       chain.Move
}
