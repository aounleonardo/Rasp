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
	isMessagesRequest, _ := regexp.MatchString("/message/", r.RequestURI)
	if isMessagesRequest {
		statusPacket := constructStatusPacket(r.RequestURI)
		return json.Marshal(waitForMessages(conn, statusPacket))
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
	status message.StatusPacket,
) []message.PeerMessages {
	response := &message.MessagesResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			Messages: &message.MessagesRequest{Status: status},
		},
		response,
	)
	return response.Messages
}

func constructStatusPacket(uri string) message.StatusPacket {
	status := strings.TrimSuffix(
		strings.TrimPrefix(uri, "/message/"),
		"/",
	)
	if len(status) == 0 {
		return message.StatusPacket{Want: nil}
	}
	wants := strings.Split(status, ";")
	statusPacket := make([]message.PeerStatus, len(wants))
	for index, peer := range wants {
		want := strings.Split(peer, ":")
		nextID, err := strconv.Atoi(want[1])
		if err != nil {
			nextID = 1
		}
		statusPacket[index] = message.PeerStatus{
			Identifier: want[0],
			NextID:     uint32(nextID),
		}
	}
	return message.StatusPacket{Want: statusPacket}
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
