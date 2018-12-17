package message

type RaspRequest struct {
	Identifier  uint64
	Bet         uint32
	Destination string
	Origin      string
	Signature   uint64
}

type RaspResponse struct {
	Destination string
	Origin      string
	Signature   uint64
}

type RaspAttack struct {
	Destination string
	Origin      string
	SignedMove  uint64
}

type RaspDefence struct {
	Destination string
	Origin      string
	SignedMove  uint64
}
