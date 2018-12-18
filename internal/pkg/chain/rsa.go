package chain

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"github.com/dedis/protobuf"
)

const keySize = 1024
const hashFunctionType = crypto.SHA256
var hashFunction = sha256.Sum256
type uid = uint64
type bet = uint32

type RequestSignature struct {
	Identifier uid
	Bet bet
}

type ResponseSignature struct {
	Identifier uint64
}

type AttackSignature1 struct {
	Identifier uid
	Bet uint32
}

type AttackSignature2 struct {
	Identifier uid
	Move int
	Nonce uint64
}

type DefenceSignature struct {
	Identifier uid
	Move int
}

type RevealSignature struct {
	Identifier uid
	Move int
	Nonce uint64
}

func GenerateKeys() (private *rsa.PrivateKey, public *rsa.PublicKey, err error) {

	reader := rand.Reader

	private, err = rsa.GenerateKey(reader, keySize)

	if err == nil {
		public = &rsa.PublicKey{N: private.N, E: private.E}
	}

	return

}

func sign(private *rsa.PrivateKey, enc []byte) (sig []byte, err error) {

	hash := hashFunction(enc)

	sig, err = rsa.SignPKCS1v15(rand.Reader, private, hashFunctionType, hash[:])

	return

}

func verify(public *rsa.PublicKey, enc []byte, sig []byte) (ok bool) {

	hash := hashFunction(enc)

	ok =  rsa.VerifyPKCS1v15(public, hashFunctionType, hash[:], sig) == nil

	return

}

func SignRequest(private *rsa.PrivateKey, id uid, b bet) (sig []byte, err error) {

	req := &RequestSignature{id, b}

	enc, err := protobuf.Encode(req)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return

}

func VerifyRequest(public *rsa.PublicKey, id uid, b bet, sig []byte) (ok bool, err error) {

	req := &RequestSignature{id, b}

	enc, err := protobuf.Encode(req)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}

func SignResponse(private *rsa.PrivateKey, id uid) (sig []byte, err error) {

	res := &ResponseSignature{id}

	enc, err := protobuf.Encode(res)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return
}

func VerifyResponse(public *rsa.PublicKey, id uid, sig []byte) (ok bool, err error) {

	req := &ResponseSignature{id}

	enc, err := protobuf.Encode(req)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}

func SignAttack(private *rsa.PrivateKey, id uid, b bet,
	move int, nonce uint64)(sig1 []byte, sig2 []byte, err error) {

		att1 := &AttackSignature1{id, b}
		att2 := &AttackSignature2{id, move, nonce}

		enc1, err := protobuf.Encode(att1)

		if err == nil {
			sig1, err = sign(private, enc1)
		}

		if err != nil {
			return
		}

		enc2, err := protobuf.Encode(att2)

		if err == nil {
			sig2, err = sign(private, enc2)
		}

		return

}

func VerifyAttack(public *rsa.PublicKey, id uid, b bet,
	move int, nonce uint64, sig1 []byte, sig2 []byte) (ok bool, err error) {

	att1 := &AttackSignature1{id, b}
	att2 := &AttackSignature2{id, move, nonce}

	enc1, err := protobuf.Encode(att1)

	if err == nil {

		ok = verify(public, enc1, sig1)
	}

	if err != nil {
		return
	}

	enc2, err := protobuf.Encode(att2)

	if err == nil {

		ok = verify(public, enc2, sig2)
	}

	return

}

func SignDefence(private *rsa.PrivateKey, id uid, move int) (sig []byte, err error) {

	def := &DefenceSignature{id, move}

	enc, err := protobuf.Encode(def)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return

}

func VerifyDefence(public *rsa.PublicKey, id uid, move int, sig []byte) (ok bool, err error) {

	def := &DefenceSignature{id, move}

	enc, err := protobuf.Encode(def)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}

func SignReveal(private *rsa.PrivateKey, id uid, move int, nonce uint64)(sig []byte, err error) {

	rev := &RevealSignature{id, move, nonce}

	enc, err := protobuf.Encode(rev)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return

}

func VerifyReveal(public *rsa.PublicKey, id uid, move int, nonce uint64, sig []byte) (ok bool, err error) {

	rev := &RevealSignature{id, move, nonce}

	enc, err := protobuf.Encode(rev)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}