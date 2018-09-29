package main

import (
	"flag"
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
		"",
		"message to be sent",
	)
	flag.Parse()

	println(*msg)
	destinationAddr := "127.0.0.1:" + *uiPort
	conn, _ := net.Dial("udp4", destinationAddr)
	defer conn.Close()
}