package main

import (
	"fmt"
	"log"
	"net/http"
	"net"
	"github.com/dedis/protobuf"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"regexp"
	"encoding/json"
	"errors"
	"strings"
	"strconv"
)

func multiplexer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	destinationAddr, _ := net.ResolveUDPAddr(
		"udp4",
		"127.0.0.1:8080",
	)
	conn, _ := net.DialUDP("udp4", nil, destinationAddr)
	defer conn.Close()

	var ret []byte
	var err error
	switch r.Method {
	case "GET":
		ret, err = getHandler(r, conn)
	case "POST":
		ret, err = postHandler(r, conn)

	}
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, string(ret))
	}
}

func getHandler(r *http.Request, conn *net.UDPConn) ([]byte, error) {
	isIdentifierRequest, _ := regexp.MatchString("/identifier/", r.RequestURI)
	if isIdentifierRequest {
		return json.Marshal(waitForIdentifier(conn))
	}
	isPeerRequest, _ := regexp.MatchString("/peers/", r.RequestURI)
	if isPeerRequest {
		return json.Marshal(waitForPeers(conn))
	}
	isMessagesRequest, _ := regexp.MatchString("/message/*", r.RequestURI)
	if isMessagesRequest {
		start := getStartIndex(r.RequestURI)
		return json.Marshal(waitForMessages(conn, start))
	}
	return nil, errors.New("unsupported URI")
}

func postHandler(r *http.Request, conn *net.UDPConn) ([]byte, error) {
	return nil, errors.New("unsupported URI")
}

func waitForIdentifier(conn *net.UDPConn) string {
	response := &message.IdentifierResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{Identifier: &message.IdentifierRequest{}},
		response,
	)
	return response.Identifier
}

func waitForPeers(conn *net.UDPConn) []string {
	response := &message.PeersResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{Peers: &message.PeersRequest{}},
		response,
	)
	return response.Peers
}

func waitForMessages(
	conn *net.UDPConn,
	start int,
) message.MessagesResponse {
	response := &message.MessagesResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			Messages: &message.MessagesRequest{StartIndex: start},
		},
		response,
	)
	return *response
}

func getStartIndex(uri string) int {
	start := strings.TrimSuffix(
		strings.TrimPrefix(uri, "/message/"),
		"/",
	)
	if len(start) == 0 {
		return 0
	}
	nextID, err := strconv.Atoi(start)
	if err != nil {
		return 0
	}
	return nextID
}

func contactGossiper(
	conn *net.UDPConn,
	request *message.ClientPacket,
	response interface{},
) {
	bytes, _ := protobuf.Encode(request)
	conn.Write(bytes)
	for {
		bytes := make([]byte, 1024)
		_, _ = conn.Read(bytes)
		protobuf.Decode(bytes, response)
		return
	}
}

func main() {
	http.HandleFunc("/", multiplexer)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
