package chain

type RequestSignature struct {
	Hash []byte // X, B
}

type ResponseSignature struct {
	Identifier uint64
}

type AttackSignature struct {
	Identifier uint64
	Bet uint32
	Hash []byte // X, B, Move, Nonce
}

type DefenceSignature struct {
	Identifier uint64
	Move int
	Hash []byte // X, Move
}

type RevealSignature struct {
	Move int
	Nonce uint64
	Hash []byte // X, Move, Nonce
}
