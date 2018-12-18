package chain

import (
	"crypto"
	"sync"
	"math/rand"
	"time"
	"fmt"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"crypto/rsa"
	"github.com/dedis/onet/log"
	"errors"
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
type Nonce = uint64
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

type Match struct {
	Identifier  Uid
	Attacker    string
	Defender    *string
	Bet         uint32
	AttackMove  *Move
	DefenceMove *Move
	Nonce       *Nonce
	Stage       Stage
}

var raspState = struct {
	sync.Mutex
	matches  map[Uid]*Match
	proposed map[Uid]struct{}
	pending  map[Uid]struct{}
	accepted map[Uid]struct{}
	ongoing  map[Uid]struct{}
	finished map[Uid]struct{}
}{
	matches:  make(map[Uid]*Match),
	proposed: make(map[Uid]struct{}),
	pending:  make(map[Uid]struct{}),
	accepted: make(map[Uid]struct{}),
	ongoing:  make(map[Uid]struct{}),
	finished: make(map[Uid]struct{}),
}

func StartGame() *rsa.PrivateKey {
	rand.Seed(time.Now().UnixNano())
	private, public, err := GenerateKeys()
	if err != nil {
		log.Fatal("error generating keys", err.Error())
	}
	// TODO advertise public key
	time.Sleep(time.Second)
	fmt.Println("Starting Game", public)
	go Mine()
	return private
}

func createMatchUID() Uid {
	return rand.Uint64()
}

func createNonce() Nonce {
	return rand.Uint64()
}

func CreateMatch(
	destination *string,
	bet Bet,
	move Move,
	gossiper string,
	privateKey *rsa.PrivateKey,
) (request message.RaspRequest, err error) {
	/* TODO
		verify both players exists
		have enough balance
		throw otherwise
	*/
	uid := createMatchUID()
	nonce := createNonce()
	newMatch := &Match{
		Identifier:  uid,
		Attacker:    gossiper,
		Defender:    destination,
		Bet:         bet,
		AttackMove:  &move,
		DefenceMove: nil,
		Nonce:       &nonce,
		Stage:       Spawn,
	}

	raspState.Lock()
	raspState.matches[uid] = newMatch
	raspState.proposed[uid] = struct{}{}
	raspState.Unlock()

	signature, err := SignRequest(privateKey, uid, bet)
	if err != nil {
		err = errors.New(
			fmt.Sprintf("error signing request %s", err.Error()),
		)
	}

	request = message.RaspRequest{
		Identifier:  uid,
		Bet:         bet,
		Destination: destination,
		Origin:      gossiper,
		Signature:   signature,
	}
	return
}
