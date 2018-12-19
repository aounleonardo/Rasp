package chain

import (
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dedis/onet/log"
	"math/rand"
	"sync"
	"time"
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
	sync.RWMutex
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
	// TODO advertise public key, and create random identifier
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
) (request *RaspRequest, err error) {
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

	request = &RaspRequest{
		Identifier:  uid,
		Bet:         bet,
		Destination: destination,
		Origin:      gossiper,
		Signature:   signature,
	}
	return
}

func AcceptMatch(
	id Uid,
	move Move,
	gossiper string,
	privateKey *rsa.PrivateKey,
) (response *RaspResponse, err error) {
	if !isMatchPending(id) {
		err = errors.New(fmt.Sprintf("match %d is not pending", id))
		return
	}
	raspState.Lock()
	defer raspState.Unlock()

	// TODO check balances

	raspState.matches[id].DefenceMove = &move
	delete(raspState.pending, id)
	raspState.accepted[id] = struct{}{}

	// TODO put a timeout to check if it is not ongoing -> set pending again ?

	signature, err := SignResponse(privateKey, id)

	response = &RaspResponse{
		Destination: raspState.matches[id].Attacker,
		Origin:      gossiper,
		Identifier:  id,
		Signature:   signature,
	}
	return
}

func HasSeenMatch(id Uid) bool {
	raspState.RLock()
	defer raspState.RUnlock()
	_, exists := raspState.matches[id]
	return exists
}

func isMatchPending(id Uid) bool {
	raspState.RLock()
	defer raspState.RUnlock()
	_, exists := raspState.pending[id]
	return exists
}
