package chain

import "fmt"

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
	SignedMove  Signature
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
	if err != nil || !verified{
		fmt.Println("error verifying", request)
		return
	}
	raspState.Lock()
	raspState.matches[request.Identifier] = &Match{
		Identifier: request.Identifier,
		Attacker: request.Origin,
		Defender: request.Destination,
		Bet: request.Bet,
	}
	raspState.pending[request.Identifier] = struct{}{}
	raspState.Unlock()
}

func ReceiveRaspResponse(response RaspResponse) {

}

func ReceiveRaspAttack(attack RaspAttack) {

}

func ReceiveRaspDefence(defence RaspDefence) {

}
