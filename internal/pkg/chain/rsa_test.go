package chain

import (
	"crypto/x509"
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

func TestProtobufKey(t *testing.T) {
	_, public, _ := GenerateKeys()

	enc := x509.MarshalPKCS1PublicKey(public)

	dec, err := x509.ParsePKCS1PublicKey(enc)

	if err != nil {
		t.Error("error decoding key", err.Error())
		return
	}

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
