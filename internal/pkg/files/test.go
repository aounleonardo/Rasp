package files

import (
	"fmt"
	"io/ioutil"
)

func TestCombineChunksIntoFile(metakey string, filename string) error {
	_ = NewFileState(metakey, filename)
	metafile, err := ioutil.ReadFile(chunksDownloads + metakey)
	if err != nil {
		fmt.Println("error reading metafile:", err.Error())
		return err
	}
	InitFileState(metafile)
	FileStates.Lock()
	defer FileStates.Unlock()
	_, err = combineChunksIntoFile(
		FileStates.m[metakey].Chunkeys,
		filename,
	)
	if err != nil {
		fmt.Println("error combining file:", err.Error())
		return err
	}
	fmt.Printf("RECONSTRUCTED file %s\n", FileStates.m[metakey].Filename)
	delete(FileStates.m, metakey)
	return nil
}
