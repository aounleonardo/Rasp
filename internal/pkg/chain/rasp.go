package chain

import (
	"crypto"
	"sync"
	"math/rand"
	"time"
	"fmt"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
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

type Uid = uint64
type Bet = uint32
type Move = int
type Stage = int

type Player struct {
	Key     crypto.PublicKey
	Balance int64
}

type GameAction struct {
	Type       int
	Identifier Uid
	Attacker   string
	Defender   string
	Bet        Bet
	Special    []byte
}

type ChallengeState struct {
	Identifier  uint64
	Attacker    string
	Defender    *string
	Bet         uint32
	AttackMove  *Move
	DefenceMove *Move
	Nonce       *uint64
	Stage       Stage
}

var raspState = struct {
	sync.Mutex
	challenges map[Uid]*ChallengeState
	proposed   []Uid
	pending    []Uid
	accepted   []Uid
	ongoing    []Uid
	finished   []Uid
}{
	challenges: make(map[Uid]*ChallengeState),
	proposed:   make([]Uid, 0),
	pending:    make([]Uid, 0),
	accepted:   make([]Uid, 0),
	ongoing:    make([]Uid, 0),
	finished:   make([]Uid, 0),
}

func StartGame() {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Second)
	fmt.Println("Starting Game")
	go Mine()
}

func createChallengeUID() Uid {
	return rand.Uint64()
}

func CreateChallenge(
	destination *string,
	bet Bet,
	move Move,
) (request message.RaspRequest, err error) {
	return
}
