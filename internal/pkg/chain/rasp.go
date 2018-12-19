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

type Winner = int

const (
	Draw     = iota
	Attacker = iota
	Defender = iota
)

type Player struct {
	Key     crypto.PublicKey
	Balance int64
}

type GameAction struct {
	Type          int
	Identifier    uint64
	Attacker      string
	Defender      string
	Bet           uint32
	Move          int
	Nonce         uint64
	HiddenMove    []byte
	SignedSpecial []byte
}

type ChallengeState struct {
	Identifier  uint64
	Attacker    string
	Defender    *string
	Bet         uint32
	AttackMove  *int
	DefenceMove *int
	Nonce       *uint64
	HiddenMove  []byte
	Stage       int
}

func whoWon(attackerMove int, defenderMove int) Winner {
	if attackerMove == defenderMove {
		return Draw
	}
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
}
