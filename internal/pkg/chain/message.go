package chain

import (
	"fmt"
	"crypto/rsa"
	"errors"
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
	privateKey *rsa.PrivateKey,
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
	*match.Defender = response.Origin
	delete(raspState.proposed, response.Identifier)
	raspState.ongoing[response.Identifier] = struct{}{}

	if err != nil {
		return
	}

	// TODO sign attack
	action := GameAction{
		Type: Attack,
		Identifier: response.Identifier,
		Attacker: response.Destination,
		Defender: response.Origin,
		Bet: match.Bet,
	}
	publishAction(action)

	attack = &RaspAttack{
		Destination: response.Origin,
		Origin: response.Destination,
		Identifier: response.Identifier,
		Bet: match.Bet,
	}
	return
}

func ReceiveRaspAttack(
	attack RaspAttack,
	privateKey *rsa.PrivateKey,
) (defence *RaspDefence, err error) {

	return
}

func ReceiveRaspDefence(defence RaspDefence, privateKey *rsa.PrivateKey) {
	// TODO ReceiveRaspDefence
}
