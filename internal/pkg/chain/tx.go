package chain

type TxPublish struct {
	File     File
	HopLimit uint32
}

type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}
