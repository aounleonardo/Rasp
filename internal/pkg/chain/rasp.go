package chain

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dedis/onet/log"
	"math/rand"
	"sync"
	"time"
	"strconv"
)

const attackerPatience = 3 * time.Second

type Stage = int

const (
	Spawn   = iota
	Attack  = iota
	Defence = iota
	Reveal  = iota
	Cancel  = iota
)

type Move = int

const (
	Rock     = iota
	Paper    = iota
	Scissors = iota
)

type Winner = int

const (
	Draw     = iota
	Attacker = iota
	Defender = iota
)

type Uid = string
type Nonce = string
type Bet = uint32

type Player struct {
	Key     rsa.PublicKey
	Balance int64
}

type GameAction struct {
	Type          Stage
	Identifier    Uid
	Attacker      string
	Defender      string
	Bet           Bet
	Move          Move
	Nonce         Nonce
	HiddenMove    Signature
	SignedSpecial Signature
}

type Match struct {
	Identifier  Uid
	Attacker    string
	Defender    *string
	Bet         Bet
	AttackMove  *Move
	DefenceMove *Move
	Nonce       *Nonce
	HiddenMove  Signature
	Stage       Stage
}

func copyMatchUnsafe(match *Match) *Match {
	return &Match{
		Identifier:  match.Identifier,
		Attacker:    match.Attacker,
		Defender:    match.Defender,
		Bet:         match.Bet,
		AttackMove:  match.AttackMove,
		DefenceMove: match.DefenceMove,
		Nonce:       match.Nonce,
		HiddenMove:  match.HiddenMove,
		Stage:       match.Stage,
	}
}

var gossiperKey *rsa.PrivateKey

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

func getState(identifier Uid) (copy *Match, exists bool) {
	raspState.RLock()
	defer raspState.RUnlock()
	state, exists := raspState.matches[identifier]
	if exists {
		copy = copyMatchUnsafe(state)
	}
	return
}

func StartGame(gossiper string) {
	rand.Seed(time.Now().UnixNano())
	private, public, err := GenerateKeys()
	gossiperKey = private
	if err != nil {
		log.Fatal("error generating keys", err.Error())
	}
	fmt.Println("Connecting with other players, they don't seem really nice")

	time.Sleep(3 * time.Second)
	fmt.Println("Introducing myself, letting them know who's the boss!")

	publishKey(gossiper, public)

	time.Sleep(3 * time.Second)
	fmt.Println("Starting Game, gonna beat the shit out of them!")

	go Mine()
}

func publishKey(gossiper string, public *rsa.PublicKey) {
	newSpawn := GameAction{
		Type:          Spawn,
		Identifier:    createUID(),
		Attacker:      gossiper,
		SignedSpecial: encodeKey(public),
	}
	go publishAction(newSpawn)
}

func createUID() Uid {
	return strconv.FormatUint(rand.Uint64(), 10)
}

func createNonce() Nonce {
	return strconv.FormatUint(rand.Uint64(), 10)
}

func CreateMatch(
	destination *string,
	bet Bet,
	move Move,
	gossiper string,
) (request *RaspRequest, err error) {
	player, exists := getPlayer(gossiper)
	if !exists {
		err = errors.New(fmt.Sprintf(
			"%s do not exist on the current ledger",
			gossiper,
		))
		return
	}
	if !player.hasEnoughMoney(bet) {
		err = errors.New(fmt.Sprintf(
			"%s does not have enough money on the current ledger",
			gossiper,
		))
		return
	}

	uid := createUID()
	nonce := createNonce()
	hiddenMove, err := SignHiddenMove(gossiperKey, uid, move, nonce)
	if err != nil {
		fmt.Println("Error while signing hidden move", err.Error())
		return
	}

	newMatch := &Match{
		Identifier:  uid,
		Attacker:    gossiper,
		Defender:    destination,
		Bet:         bet,
		AttackMove:  &move,
		DefenceMove: nil,
		Nonce:       &nonce,
		HiddenMove:  hiddenMove,
		Stage:       Spawn,
	}

	raspState.Lock()
	raspState.matches[uid] = newMatch
	raspState.proposed[uid] = struct{}{}
	raspState.Unlock()
	if newMatch.Defender == nil {
		fmt.Printf(
			"CREATE OPEN MATCH: Attacker %s, Bet %d, UID %s, AttackMove %d\n",
			newMatch.Attacker,
			newMatch.Bet,
			newMatch.Identifier,
			*newMatch.AttackMove,
			)

	} else {
		fmt.Printf(
			"CREATE MATCH: Attacker %s, Defender %s, Bet %d, UID %s, Attack Move %d\n",
			newMatch.Attacker,
			*newMatch.Defender,
			newMatch.Bet,
			newMatch.Identifier,
			*newMatch.AttackMove,
			)

	}

	signature, err := SignRequest(gossiperKey, uid, bet)
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
) (response *RaspResponse, err error) {
	if !isMatchPending(id) {
		err = errors.New(fmt.Sprintf("match %s is not pending", id))
		return
	}
	raspState.Lock()
	defer raspState.Unlock()
	match := raspState.matches[id]

	player, playerExists := getPlayer(gossiper)
	opponent, opponentExists := getPlayer(match.Attacker)
	if !playerExists || !opponentExists {
		err = errors.New(fmt.Sprintf(
			"%s and %s do not both exist in the current ledger",
			gossiper,
			match.Attacker,
		))
		return
	}
	if !player.hasEnoughMoney(match.Bet) ||
		!opponent.hasEnoughMoney(match.Bet) {
		err = errors.New(fmt.Sprintf(
			"%s or %s do not have enough money in the current ledger",
			gossiper,
			match.Attacker,
		))
		return
	}

	match.Defender = &gossiper
	match.DefenceMove = &move
	delete(raspState.pending, id)
	raspState.accepted[id] = struct{}{}
	fmt.Printf(
		"ACCEPTING MATCH: Attacker %s,"+
			" Defender %s,"+
			" Bet %d,"+
			" UID %s,"+
			" Defense Move %d\n",
		match.Attacker,
		*match.Defender,
		match.Bet,
		match.Identifier,
		*match.DefenceMove)

	signature, err := SignResponse(gossiperKey, id)
	response = &RaspResponse{
		Destination: match.Attacker,
		Origin:      gossiper,
		Identifier:  id,
		Signature:   signature,
	}
	return
}

func GetPlayers(players *PlayersResponse) {
	blockchain.RLock()
	defer blockchain.RUnlock()
	players.Players = make(map[string]int64)
	for s, p := range blockchain.heads[blockchain.longest].players {
		players.Players[s] = p.Balance
	}

}

func GetStates(states *StateResponse) {
	raspState.RLock()
	defer raspState.RUnlock()

	states.Matches = copyMatchesUnsafe()
	states.Proposed = copyStatesUnsafe(raspState.proposed)
	states.Pending = copyStatesUnsafe(raspState.pending)
	states.Accepted = copyStatesUnsafe(raspState.accepted)
	states.Ongoing = copyStatesUnsafe(raspState.ongoing)
	states.Finished = copyStatesUnsafe(raspState.finished)
}

func copyMatchesUnsafe() map[Uid]*Match {
	copy := make(map[Uid]*Match)
	for uid, match := range raspState.matches {
		copy[uid] = match
	}
	return copy
}

func copyStatesUnsafe(state map[Uid]struct{}) map[Uid]struct{} {
	copy := make(map[string]struct{})
	for uid := range state {
		copy[uid] = struct{}{}
	}
	return copy
}

func isMatchPending(id Uid) bool {
	raspState.RLock()
	defer raspState.RUnlock()
	_, exists := raspState.pending[id]
	return exists
}

func whoWon(attackerMove int, defenderMove int) Winner {
	switch attackerMove {
	case Rock:
		switch defenderMove {
		case Paper:
			return Defender
		case Scissors:
			return Attacker
		}
	case Paper:
		switch defenderMove {
		case Rock:
			return Attacker
		case Scissors:
			return Defender
		}
	case Scissors:
		switch defenderMove {
		case Rock:
			return Defender
		case Paper:
			return Attacker
		}
	}
	return Draw
}

func RaspStateUpdateUnsafe(newLedger ledger) {
	raspState.Lock()
	defer raspState.Unlock()
	for x, myMatch := range raspState.matches {
		match, exists := newLedger.matches[x]
		if exists && match.Stage > Defence {
			delete(raspState.pending, x)
			delete(raspState.proposed, x)
			delete(raspState.accepted, x)
			delete(raspState.ongoing, x)
			raspState.finished[x] = struct{}{}
		} else if myMatch.Stage > Spawn {
			delete(raspState.finished, x)
			raspState.ongoing[x] = struct{}{}
		}
	}
}
