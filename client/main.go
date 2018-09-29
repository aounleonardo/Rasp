package main

import (
	"flag"
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/dedis/protobuf"
	"fmt"
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

	println(*msg)
	destinationAddr := "127.0.0.1:" + *uiPort
	conn, _ := net.Dial("udp4", destinationAddr)

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

	defer conn.Close()
}