package chain

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"
)

type Signature = []byte

type RaspRequest struct {
	Identifier  Uid
	Bet         Bet
	Destination *string
	Origin      string
	Signature   Signature
}

type RaspResponse struct {
	Destination string
	Origin      string
	Identifier  Uid
	Signature   Signature
}

type RaspAttack struct {
	Destination string
	Origin      string
	Identifier  Uid
	Bet         Bet
	SignedBet   Signature
	HiddenMove  Signature
}

type RaspDefence struct {
	Destination string
	Origin      string
	Identifier  Uid
	Move        Move
	SignedMove  Signature
}

type CreateMatchRequest struct {
	Destination *string
	Bet         Bet
	Move        Move
}

type AcceptMatchRequest struct {
	Identifier Uid
	Move       Move
}

type PlayersRequest struct {}

type PlayersResponse struct {
	Players map[string]int64
}

type StateRequest struct {}

type StateResponse struct {
	Matches  map[Uid]*Match
	Proposed map[Uid]struct{}
	Pending  map[Uid]struct{}
	Accepted map[Uid]struct{}
	Ongoing  map[Uid]struct{}
	Finished map[Uid]struct{}
}

func ReceiveRaspRequest(request RaspRequest) {
	opponent, exists := getPlayer(request.Origin)
	if !exists {
		return
	}
	verified, err := VerifyRequest(
		&opponent.Key,
		request.Identifier,
		request.Bet,
		request.Signature,
	)
	if err != nil || !verified {
		fmt.Println("error verifying", request)
		return
	}
	raspState.Lock()
	raspState.matches[request.Identifier] = &Match{
		Identifier: request.Identifier,
		Attacker:   request.Origin,
		Defender:   request.Destination,
		Bet:        request.Bet,
	}
	raspState.pending[request.Identifier] = struct{}{}
	raspState.Unlock()
}

func ReceiveRaspResponse(
	response RaspResponse,
) (attack *RaspAttack, err error) {
	opponent, exists := getPlayer(response.Origin)
	if !exists {
		err = errors.New(
			fmt.Sprintf("%s does not exist", response.Origin),
		)
		return
	}
	verified, err := VerifyResponse(
		&opponent.Key,
		response.Identifier,
		response.Signature,
	)
	if err != nil || !verified {
		err = errors.New(fmt.Sprintf(
			"error verifying response %d",
			response.Identifier,
		))
		return
	}

	raspState.Lock()
	defer raspState.Unlock()
	match, exists := raspState.matches[response.Identifier]
	if !exists ||
		(match.Defender != nil && *match.Defender != response.Origin) {
		err = errors.New(fmt.Sprintf(
			"match %d with opponent %s does not exist",
			response.Identifier,
			response.Origin,
		))
		return
	}
	if _, isProposed := raspState.proposed[response.Identifier]; !isProposed {
		err = errors.New(fmt.Sprintf(
			"match %d is not proposed",
			response.Identifier,
		))
		return
	}
	match.Defender = &response.Origin
	delete(raspState.proposed, response.Identifier)
	raspState.ongoing[response.Identifier] = struct{}{}

	if err != nil {
		return
	}

	signature, err := SignAttack(
		gossiperKey,
		match.Identifier,
		match.Bet,
	)
	action := GameAction{
		Type:          Attack,
		Identifier:    response.Identifier,
		Attacker:      response.Destination,
		Defender:      response.Origin,
		Bet:           match.Bet,
		HiddenMove:    match.HiddenMove,
		SignedSpecial: signature,
	}
	publishAction(action)
	go waitForDefenceTimeout(response.Identifier, gossiperKey)

	attack = &RaspAttack{
		Destination: response.Origin,
		Origin:      response.Destination,
		Identifier:  response.Identifier,
		Bet:         match.Bet,
		SignedBet:   signature,
		HiddenMove:  match.HiddenMove,
	}
	match.Stage = Attack
	return
}

func waitForDefenceTimeout(identifier Uid, privateKey *rsa.PrivateKey) {
	time.Sleep(attackerPatience)
	if match, exists := getState(identifier); exists && match.Stage < Defence {
		cancel, err := createCancel(match, privateKey)
		if err != nil {
			fmt.Println("cannot send cancel for,", identifier, err.Error())
			return
		}
		publishAction(cancel)
	}
}

func createReveal(
	match *Match,
	key *rsa.PrivateKey,
	defence GameAction,
) (action GameAction, err error) {

	signature, err := SignReveal(
		key,
		match.Identifier,
		*match.AttackMove,
		*match.Nonce,
		defence.SignedSpecial,
	)
	if err != nil {
		return
	}
	action = GameAction{
		Type:          Reveal,
		Identifier:    match.Identifier,
		Attacker:      match.Attacker,
		Defender:      *match.Defender,
		Move:          *match.AttackMove,
		Nonce:         *match.Nonce,
		HiddenMove:    defence.SignedSpecial,
		SignedSpecial: signature,
	}
	return
}

func createCancel(
	match *Match,
	key *rsa.PrivateKey,
) (action GameAction, err error) {
	signature, err := SignCancel(key, match.Identifier)
	if err != nil {
		return
	}
	action = GameAction{
		Type:          Cancel,
		Identifier:    match.Identifier,
		Attacker:      match.Attacker,
		Defender:      *match.Defender,
		Bet:           match.Bet,
		SignedSpecial: signature,
	}
	return
}

func ReceiveRaspAttack(attack RaspAttack) (defence *RaspDefence, err error) {
	opponent, opponentExists := getPlayer(attack.Origin)
	if !opponentExists {
		err = errors.New(
			fmt.Sprintf("%s does not exist", attack.Origin),
		)
		return
	}
	attackerPublic := opponent.Key
	ok, err := VerifyAttack(
		&attackerPublic,
		attack.Identifier,
		attack.Bet,
		attack.SignedBet,
	)
	if !ok {
		err = errors.New(
			fmt.Sprintf("Unable to verify the signedBet from %s",
				attack.Origin),
		)
		return
	}
	raspState.RLock()
	_, ok = raspState.accepted[attack.Identifier]
	raspState.RUnlock()
	if !ok {
		err = errors.New(
			fmt.Sprintf("Unable to find the corresponding match in Accepted %d",
				attack.Identifier),
		)
		return
	}
	raspState.Lock()
	delete(raspState.accepted, attack.Identifier)
	raspState.ongoing[attack.Identifier] = struct{}{}
	match := raspState.matches[attack.Identifier]
	defenseMove := match.DefenceMove
	match.HiddenMove = attack.HiddenMove
	defenceSpecial, err := SignDefence(
		gossiperKey,
		attack.Identifier,
		*defenseMove)
	if err != nil {
		err = errors.New(
			fmt.Sprintf("Unable to sign the defence move %d",
				defenseMove),
		)
		return
	}
	match.Stage = Defence
	raspState.Unlock()

	action := GameAction{
		Type:          Defence,
		Identifier:    attack.Identifier,
		Attacker:      attack.Origin,
		Defender:      attack.Destination,
		Bet:           attack.Bet,
		Move:          *defenseMove,
		SignedSpecial: defenceSpecial,
	}
	publishAction(action)

	defence = &RaspDefence{
		Destination: attack.Origin,
		Origin:      attack.Destination,
		Identifier:  attack.Identifier,
		Move:        *defenseMove,
		SignedMove:  defenceSpecial,
	}
	return
}

func ReceiveRaspDefence(defence RaspDefence) {
	defender, defenderExists := getPlayer(defence.Origin)
	if !defenderExists {
		fmt.Printf("RaspDefence error: %s does not exist\n", defence.Origin)
		return
	}
	defenderPublic := defender.Key
	ok, err := VerifyDefence(&defenderPublic,
		defence.Identifier,
		defence.Move,
		defence.SignedMove,
	)
	if !ok {
		fmt.Println("RaspDefence error: Unable to verify signed move", err)
		return
	}
	raspState.RLock()
	_, ok = raspState.ongoing[defence.Identifier]
	raspState.RUnlock()
	if !ok {
		fmt.Printf("Unable to find the corresponding match in Ongoing %d\n",
			defence.Identifier)
		return
	}
	raspState.Lock()
	defer raspState.Unlock()
	match := raspState.matches[defence.Identifier]
	attackMove := match.AttackMove
	nonce := match.Nonce

	match.DefenceMove = &defence.Move

	revealSign, err := SignReveal(
		gossiperKey,
		defence.Identifier,
		*attackMove,
		*nonce,
		defence.SignedMove,
	)
	if err != nil {
		fmt.Println("Unable to sign the reveal")
		return
	}

	action := GameAction{
		Type:          Reveal,
		Identifier:    defence.Identifier,
		Attacker:      defence.Destination,
		Defender:      defence.Origin,
		Move:          *attackMove,
		Nonce:         *nonce,
		HiddenMove:    defence.SignedMove,
		SignedSpecial: revealSign,
	}
	publishAction(action)
	match.Stage = Reveal
}
