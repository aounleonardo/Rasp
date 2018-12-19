package chain

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"github.com/dedis/protobuf"
)

const keySize = 1024
const hashFunctionType = crypto.SHA256

var hashFunction = sha256.Sum256

type RequestSignature struct {
	Identifier Uid
	Bet        Bet
}

type ResponseSignature struct {
	Identifier uint64
}

type AttackSignature struct {
	Identifier Uid
	Bet        uint32
}

type HiddenMoveSignature struct {
	Identifier Uid
	Move       int
	Nonce      uint64
}

type DefenceSignature struct {
	Identifier Uid
	Move       int
}

type RevealSignature struct {
	Identifier Uid
	Move       int
	Nonce      uint64
}

func GenerateKeys() (private *rsa.PrivateKey, public *rsa.PublicKey, err error) {

	reader := rand.Reader

	private, err = rsa.GenerateKey(reader, keySize)

	if err == nil {
		public = &rsa.PublicKey{N: private.N, E: private.E}
	}

	return

}

func encodeKey(public *rsa.PublicKey) []byte {
	return x509.MarshalPKCS1PublicKey(public)
}

func decodeKey(bytes []byte) *rsa.PublicKey {
	key, _ := x509.ParsePKCS1PublicKey(bytes)
	return key
}

func sign(private *rsa.PrivateKey, enc []byte) (sig []byte, err error) {

	hash := hashFunction(enc)

	sig, err = rsa.SignPKCS1v15(rand.Reader, private, hashFunctionType, hash[:])

	return

}

func verify(public *rsa.PublicKey, enc []byte, sig []byte) (ok bool) {

	hash := hashFunction(enc)

	ok = rsa.VerifyPKCS1v15(public, hashFunctionType, hash[:], sig) == nil

	return

}

func SignRequest(private *rsa.PrivateKey, id Uid, b Bet) (sig []byte, err error) {

	req := &RequestSignature{id, b}

	enc, err := protobuf.Encode(req)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return

}

func VerifyRequest(public *rsa.PublicKey, id Uid, b Bet, sig []byte) (ok bool, err error) {

	req := &RequestSignature{id, b}

	enc, err := protobuf.Encode(req)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}

func SignResponse(private *rsa.PrivateKey, id Uid) (sig []byte, err error) {

	res := &ResponseSignature{id}

	enc, err := protobuf.Encode(res)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return
}

func VerifyResponse(public *rsa.PublicKey, id Uid, sig []byte) (ok bool, err error) {

	req := &ResponseSignature{id}

	enc, err := protobuf.Encode(req)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}

func SignHiddenMove(
	private *rsa.PrivateKey,
	id Uid,
	move int,
	nonce uint64,
) (hiddenMove []byte, err error) {

	signature := &HiddenMoveSignature{Identifier: id, Move: move, Nonce: nonce}
	encoding, err := protobuf.Encode(signature)

	if err != nil {
		return
	}

	hiddenMove, err = sign(private, encoding)
	return
}

func SignAttack(
	private *rsa.PrivateKey,
	id Uid,
	b Bet,
) (sig []byte, err error) {

	att := &AttackSignature{Identifier: id, Bet: b}

	enc, err := protobuf.Encode(att)

	if err != nil {
		return
	}

	sig, err = sign(private, enc)

	return
}

func VerifyHiddenMove(
	public *rsa.PublicKey,
	id Uid,
	move int,
	nonce uint64,
	signature []byte,
) (ok bool, err error) {

	hiddenMove := &HiddenMoveSignature{Identifier: id, Move: move, Nonce: nonce}
	encoding, err := protobuf.Encode(hiddenMove)

	if err != nil {
		return
	}

	ok = verify(public, encoding, signature)

	return
}

func VerifyAttack(
	public *rsa.PublicKey,
	id Uid,
	b Bet,
	sig []byte,
) (ok bool, err error) {

	att := &AttackSignature{id, b}

	enc, err := protobuf.Encode(att)

	if err != nil {
		return
	}

	ok = verify(public, enc, sig)

	return
}

func SignDefence(private *rsa.PrivateKey, id Uid, move int) (sig []byte, err error) {

	def := &DefenceSignature{id, move}

	enc, err := protobuf.Encode(def)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return

}

func VerifyDefence(public *rsa.PublicKey, id Uid, move int, sig []byte) (ok bool, err error) {

	def := &DefenceSignature{id, move}

	enc, err := protobuf.Encode(def)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}

func SignReveal(private *rsa.PrivateKey, id Uid, move int, nonce uint64) (sig []byte, err error) {

	rev := &RevealSignature{id, move, nonce}

	enc, err := protobuf.Encode(rev)

	if err == nil {

		sig, err = sign(private, enc)

	}

	return

}

func VerifyReveal(public *rsa.PublicKey, id Uid, move int, nonce uint64, sig []byte) (ok bool, err error) {

	rev := &RevealSignature{id, move, nonce}

	enc, err := protobuf.Encode(rev)

	if err == nil {

		ok = verify(public, enc, sig)

	}

	return

}


// TODO could change name, I simply call signResponse bc it's the same code but I didn't want to break calls
func SignCancel(private *rsa.PrivateKey, id Uid) ([]byte, error) {

	return SignResponse(private, id)
}

func VerifyCancel(public *rsa.PublicKey, id Uid, sig []byte) (bool, error) {

	return VerifyResponse(public, id, sig)

}
