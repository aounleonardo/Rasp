package chain

import (
	"math/rand"
	"testing"
)

func TestSignRequest(t *testing.T) {

	id := rand.Uint64()
	b := rand.Uint32()

	private, public, err := GenerateKeys()

	if err != nil {
		t.Error("Error in key generation")
	}

	sig, err := SignRequest(private, id, b)

	if err != nil {
		t.Error("Error during signing")
	}

	ok, err := VerifyRequest(public, id, b, sig)

	if err != nil {
		t.Error("Error during verifying")
	}

	if !ok {
		t.Error("Verify failed")
	}

}

func TestSignAttack(t *testing.T) {

	id := rand.Uint64()
	b := rand.Uint32()
	move := 0
	nonce := rand.Uint64()

	private, public, err := GenerateKeys()

	if err != nil {
		t.Error("Error in key generation")
	}

	sig1, sig2, err := SignAttack(private, id, b, move, nonce)

	if err != nil {
		t.Error("Error during signing")
	}

	ok, err := VerifyAttack(public, id, b, move, nonce, sig1, sig2)

	if err != nil {
		t.Error("Error during verifying")
	}

	if !ok {
		t.Error("Verify failed")
	}

}
