package chain

import (
	"crypto/sha256"
	"encoding/binary"
)

func (b *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])
	h.Write(b.Nonce[:])
	binary.Write(h, binary.LittleEndian,
		uint32(len(b.Transactions)))
	for _, t := range b.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint64(t.Action.Identifier))
	binary.Write(h, binary.LittleEndian, uint32(t.Action.Type))
	h.Write([]byte(t.Action.Attacker))
	h.Write([]byte(t.Action.Defender))
	binary.Write(h, binary.LittleEndian, uint32(t.Action.Bet))
	h.Write(t.Action.Special)
	copy(out[:], h.Sum(nil))
	return

}
