package main

import (
	"flag"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/dedis/protobuf"
	"fmt"
	"net"
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
	flag.Parse()

	destinationAddr, _ := net.ResolveUDPAddr(
		"udp4",
		"127.0.0.1:" + *uiPort,
	)
	conn, _ := net.DialUDP("udp4", nil, destinationAddr)

	var clientPacket message.ClientPacket
	if len(*dest) > 0 {
		clientPacket = message.ClientPacket{
			SendPrivate: &message.PrivatePutRequest{
				Contents:    *msg,
				Destination: *dest,
			},
		}
	} else if len(*file) > 0 {

	} else {
		clientPacket = message.ClientPacket{
			Rumor: &message.RumorRequest{Contents: *msg},
		}
	}

	bytes, err := protobuf.Encode(&clientPacket)
	if err != nil {
		fmt.Println("Protobuf error:", err, "while encoding:", clientPacket)
		return
	}

	_, sendErr := conn.Write(bytes)
	if sendErr != nil {
		fmt.Println("Error while sending packet from client", err)
	}
	fmt.Println("Sent:", clientPacket.Rumor.Contents)

	defer conn.Close()
}
