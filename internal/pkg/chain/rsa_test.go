package chain

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

func TestSignRequest(t *testing.T) {

	id := createUID()
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

	id := createUID()
	b := rand.Uint32()
	move := 0
	nonce := createNonce()

	private, public, err := GenerateKeys()

	if err != nil {
		t.Error("Error in key generation")
	}

	sig, err := SignAttack(private, id, b)

	if err != nil {
		t.Error("Error while signing attack")
	}

	hiddenMove, err := SignHiddenMove(private, id, move, nonce)

	if err != nil {
		t.Error("Error while signing hiddenMove")
	}

	ok, err := VerifyAttack(public, id, b, sig)

	if err != nil {
		t.Error("Error while verifying attack")
	}

	ok, err = VerifyHiddenMove(public, id, move, nonce, hiddenMove)

	if err != nil {
		t.Error("Error while verifying hiddenMove")
	}

	if !ok {
		t.Error("Verify failed")
	}

}

func TestKeyEncoding(t *testing.T) {
	_, public, _ := GenerateKeys()

	enc := encodeKey(public)

	dec := decodeKey(enc)

	if dec.Size() != public.Size() {
		t.Error("not same size", public.Size(), dec.Size())
	}

	if dec.N.Cmp(public.N) != 0 {
		t.Error("not same N", *public.N, *dec.N)
	}

	if dec.E != public.E {
		t.Error("not same E", public.E, dec.E)
	}
}

func TestRequestVsAttack(t *testing.T) {
	privateKey, _, _ := GenerateKeys()
	uid := createUID()
	bet := uint32(10)

	request, _ := SignRequest(privateKey, uid, bet)
	attack, _ := SignAttack(privateKey, uid, bet)

	if bytes.Equal(request, attack) {
		t.Error("same signature")
	}
	t.Log(fmt.Sprintf("request: %x", request))
	t.Log(fmt.Sprintf("attack: %x", attack))
}
