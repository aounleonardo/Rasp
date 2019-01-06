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

type CancelMatchRequest struct{
	Identifier Uid
}

type PlayersRequest struct {}

type PlayersResponse struct {
	Players map[string]int64
}

type StateRequest struct {}

type StateResponse struct {
	Matches  map[string]*Match
	Proposed map[string]struct{}
	Pending  map[string]struct{}
	Accepted map[string]struct{}
	Ongoing  map[string]struct{}
	Finished map[string]struct{}
}

func ReceiveRaspRequest(request RaspRequest) {
	destination := "open"
	if request.Destination != nil {
		destination = *request.Destination
	}
	fmt.Printf(
		"RECEIVED RASP REQUEST: Attacker %s, Defender %s, Bet %d, UID %s\n",
		request.Origin,
		destination,
		request.Bet,
		request.Identifier,
	)
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
	defer raspState.Unlock()
	if _, proposedIt := raspState.proposed[request.Identifier]; proposedIt {
		fmt.Println("error received proposed request", request)
		return
	}
	raspState.matches[request.Identifier] = &Match{
		Identifier: request.Identifier,
		Attacker:   request.Origin,
		Defender:   request.Destination,
		Bet:        request.Bet,
	}
	raspState.pending[request.Identifier] = struct{}{}
}

func ReceiveRaspResponse(
	response RaspResponse,
) (attack *RaspAttack, err error) {
	fmt.Printf(
		"RECEIVED RASP RESPONSE: Potential Defender %s, Attacker %s, UID %s\n",
		response.Origin,
		response.Destination,
		response.Identifier,
	)
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
			"error verifying response %s",
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
			"match %s with opponent %s does not exist",
			response.Identifier,
			response.Origin,
		))
		return
	}
	if _, isProposed := raspState.proposed[response.Identifier]; !isProposed {
		err = errors.New(fmt.Sprintf(
			"match %s is not proposed",
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
	fmt.Printf(
		"RECEIVED RASP ATTACK: Attacker %s, Defender %s, Bet %d, UID %s\n",
		attack.Origin,
		attack.Destination,
		attack.Bet,
		attack.Identifier,
	)
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
			fmt.Sprintf("Unable to find the corresponding match in Accepted %s",
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
	fmt.Printf("RECEIVED RASP DEFENCE: Attacker %s, Defender %s, Move %d, UID %s\n",
		defence.Destination,
		defence.Origin,
		defence.Move,
		defence.Identifier,
	)
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
		fmt.Printf("Unable to find the corresponding match in Ongoing %s\n",
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
