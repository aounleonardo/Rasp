package message

type Signature = []byte

type RaspRequest struct {
	Identifier  uint64
	Bet         uint32
	Destination *string
	Origin      string
	Signature   Signature
}

type RaspResponse struct {
	Destination string
	Origin      string
	Signature   Signature
}

type RaspAttack struct {
	Destination string
	Origin      string
	SignedMove  Signature
}

type RaspDefence struct {
	Destination string
	Origin      string
	SignedMove  Signature
}
