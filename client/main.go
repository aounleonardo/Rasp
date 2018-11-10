package main

import (
	"flag"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/dedis/protobuf"
	"fmt"
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"io/ioutil"
	"bytes"
	"os"
)

func main() {
	uiPort := flag.String(
		"UIPort",
		"8080",
		"port for the UI client (\"default 8080\")",
	)
	msg := flag.String(
		"msg",
		"Leo",
		"message to be sent",
	)
	dest := flag.String(
		"dest",
		"",
		"destination for the private message",
	)
	file := flag.String(
		"file",
		"",
		"file to be indexed by the gossiper",
	)
	request := flag.String(
		"request",
		"",
		"request a chunk or metafile of this hash",
	)
	test := flag.String(
		"test",
		"",
		"dev only",
	)
	flag.Parse()

	destinationAddr, _ := net.ResolveUDPAddr(
		"udp4",
		"127.0.0.1:" + *uiPort,
	)
	conn, _ := net.DialUDP("udp4", nil, destinationAddr)

	var clientPacket message.ClientPacket
	if len(*test) > 0 {
		switch *test {
		case "reconstruct":
			clientPacket = message.ClientPacket{
				TestReconstruct: &message.TestFileReconstructRequest{
					Metahash: files.KeyToHash(*request),
					Filename: *file,
				},
			}
		}
	} else if len(*file) > 0 {
		if len(*request) > 0 && len(*dest) > 0 {
			clientPacket = message.ClientPacket{
				Download: &message.FileDownloadRequest{
					Name:     *file,
					Metahash: files.KeyToHash(*request),
					Origin:   *dest,
				},
			}
		} else {
			clientPacket = message.ClientPacket{FileShare: shareFile(*file)}
		}
	} else if len(*dest) > 0 {
		clientPacket = message.ClientPacket{
			SendPrivate: &message.PrivatePutRequest{
				Contents:    *msg,
				Destination: *dest,
			},
		}
	} else {
		clientPacket = message.ClientPacket{
			Rumor: &message.RumorRequest{Contents: *msg},
		}
	}

	buf, err := protobuf.Encode(&clientPacket)
	if err != nil {
		fmt.Println("Protobuf error:", err, "while encoding:", clientPacket)
		return
	}

	_, sendErr := conn.Write(buf)
	if sendErr != nil {
		fmt.Println("Error while sending packet from client", err)
	}

	defer conn.Close()
}

func shareFile(filename string) *message.FileShareRequest {
	file, err := ioutil.ReadFile(files.SharedFiles + filename)
	if err != nil {
		a, _ := os.Getwd()
		fmt.Println("error reading file", filename, err.Error(), a)
	}
	buf := bytes.NewBuffer(file)
	if buf == nil {
		fmt.Println("error building buffer for file", filename)
	}
	request, response := files.ShareFile(*buf, filename)
	if err != nil || len(response.Metakey) == 0 {
		fmt.Println("error sharing file", err.Error())
	}
	fmt.Println(response.Metakey)
	return request
}
