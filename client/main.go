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
	flag.Parse()

	destinationAddr, _ := net.ResolveUDPAddr(
		"udp4",
		"127.0.0.1:" + *uiPort,
	)
	conn, _ := net.DialUDP("udp4", nil, destinationAddr)

	clientPacket := &message.ClientPacket{Message: *msg}
	bytes, err := protobuf.Encode(clientPacket)
	if err != nil {
		fmt.Println("Protobuf error:", err, "while encoding:", clientPacket)
		return
	}

	_, sendErr := conn.Write(bytes)
	if sendErr != nil {
		fmt.Println("Error while sending packet from client", err)
	}
	fmt.Println("Sent:", clientPacket.Message)

	defer conn.Close()
}
