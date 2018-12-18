package chain

import "crypto"

const (
	Spawn   = iota
	Attack  = iota
	Defence = iota
	Reveal  = iota
	Cancel  = iota
)

const (
	Rock     = iota
	Paper    = iota
	Scissors = iota
)

type Player struct {
	Key     crypto.PublicKey
	Balance int64
}

type GameAction struct {
	Type       int
	Identifier uint64
	Attacker   string
	Defender   string
	Bet        uint32
	Special    []byte
}

type ChallengeState struct {
	Identifier  uint64
	Attacker    string
	Defender    *string
	Bet         uint32
	AttackMove  *int
	DefenceMove *int
	Nonce       *uint64
	Stage       int
}