package gossip

type File struct {
	Name     string
	Size     uint32
	Metafile []byte
	Metahash []byte
}
