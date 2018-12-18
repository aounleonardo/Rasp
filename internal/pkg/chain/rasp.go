package chain

import (
	"crypto"
	"sync"
)

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

type uid = uint64
type bet = uint32

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

var raspState = struct {
	sync.Mutex
	challenges map[uid]*ChallengeState
	proposed   []uid
	pending    []uid
	accepted   []uid
	ongoing    []uid
	finished   []uid
}{
	challenges: make(map[uid]*ChallengeState),
	proposed:   make([]uid, 0),
	pending:    make([]uid, 0),
	accepted:   make([]uid, 0),
	ongoing:    make([]uid, 0),
	finished:   make([]uid, 0),
}
