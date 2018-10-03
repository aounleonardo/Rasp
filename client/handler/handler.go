package main

import (
	"fmt"
	"log"
	"net/http"
	"net"
	"github.com/aounleonardo/Peerster/internal/pkg/requests"
	"github.com/dedis/protobuf"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	destinationAddr, _ := net.ResolveUDPAddr(
		"udp4",
		"127.0.0.1:8080",
	)
	conn, _ := net.DialUDP("udp4", nil, destinationAddr)
	defer conn.Close()

	fmt.Println("received request")

	identifier := waitForIdentifier(conn)

	fmt.Fprintf(w, "%s", identifier)
}

func waitForIdentifier(conn *net.UDPConn) string {
	request := &message.ClientPacket{Identifier:&requests.IdentifierRequest{}}
	bytes, _ := protobuf.Encode(request)
	conn.Write(bytes)
	for {
		bytes := make([]byte, 1024)
		_, _ = conn.Read(bytes)
		response := &requests.IdentifierResponse{}
		protobuf.Decode(bytes, response)
		fmt.Println(response.Identifier)
		return response.Identifier
	}
}

func main() {
	http.HandleFunc("/identifier", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}