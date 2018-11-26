package main

import (
	"fmt"
	"log"
	"net/http"
	"net"
	"github.com/dedis/protobuf"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"regexp"
	"encoding/json"
	"errors"
	"strings"
	"strconv"
	"bytes"
	"io"
	"flag"
)

var port string
var gossiperPort string

func multiplexer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	destinationAddr, _ := net.ResolveUDPAddr(
		"udp4",
		"127.0.0.1:"+gossiperPort,
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
	isIdentifierRequest, _ := regexp.MatchString(
		"/identifier/",
		r.RequestURI,
	)
	if isIdentifierRequest {
		return json.Marshal(waitForIdentifier(conn))
	}
	isPeerRequest, _ := regexp.MatchString("/peers/", r.RequestURI)
	if isPeerRequest {
		return json.Marshal(waitForPeers(conn))
	}
	isMessagesRequest, _ := regexp.MatchString(
		"/message/*",
		r.RequestURI,
	)
	if isMessagesRequest {
		start := getStartIndex(r.RequestURI)
		return json.Marshal(waitForMessages(conn, start))
	}
	if isChatRequest, _ :=
		regexp.MatchString("/chats/", r.RequestURI);
		isChatRequest {
		return json.Marshal(waitForChats(conn))
	}
	if isPrivateMessageRequest, _ :=
		regexp.MatchString("/pm/*/*/*/", r.RequestURI);
		isPrivateMessageRequest {
		partner, unordered, ordered := getPrivateIndexes(r.RequestURI)
		return json.Marshal(waitForPrivates(conn, partner, unordered, ordered))
	}
	return nil, errors.New("unsupported URI")
}

func postHandler(r *http.Request, conn *net.UDPConn) ([]byte, error) {
	if isMessagesRequest, _ :=
		regexp.MatchString("/message/", r.RequestURI);
		isMessagesRequest {
		return json.Marshal(readMessage(conn, r))
	}
	if isPeerRequest, _ := regexp.MatchString("/peers/", r.RequestURI);
		isPeerRequest {
		return json.Marshal(addPeer(conn, r))
	}
	if isPrivateMessageRequest, _ :=
		regexp.MatchString("/pm/", r.RequestURI);
		isPrivateMessageRequest {
		return json.Marshal(sendPrivateMessage(conn, r))
	}
	if isFileShareRequest, _ :=
		regexp.MatchString("/share-file/", r.RequestURI);
		isFileShareRequest {
		return json.Marshal(shareFile(conn, r))
	}
	if isDownloadRequest, _ :=
		regexp.MatchString("/download-file/", r.RequestURI);
		isDownloadRequest {
		return json.Marshal(downloadFile(conn, r))
	}
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

func waitForChats(conn *net.UDPConn) []string {
	response := &message.ChatsResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{Chats: &message.ChatsRequest{}},
		response,
	)
	return response.Origins
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

func waitForPrivates(
	conn *net.UDPConn,
	partner string,
	unorderedStart int,
	orderedStart int,
) message.PrivateGetResponse {
	response := &message.PrivateGetResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			GetPrivate: &message.PrivateGetRequest{
				Partner:        partner,
				UnorderedIndex: unorderedStart,
				OrderedIndex:   orderedStart,
			},
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

func getPrivateIndexes(uri string) (string, int, int) {
	trimmed := strings.TrimSuffix(
		strings.TrimPrefix(uri, "/pm/"),
		"/",
	)
	if len(trimmed) == 0 {
		return "", 0, 0
	}
	indexes := strings.Split(trimmed, "/")
	if len(indexes) < 3 {
		return "", 0, 0
	}
	unordered, err0 := strconv.Atoi(indexes[1])
	ordered, err1 := strconv.Atoi(indexes[2])
	if err0 != nil || err1 != nil {
		return "", 0, 0
	}
	return indexes[0], unordered, ordered
}

func readMessage(
	conn *net.UDPConn,
	r *http.Request,
) message.ValidationResponse {
	decoder := json.NewDecoder(r.Body)
	var s string
	err := decoder.Decode(&s)
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	response := &message.ValidationResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{Rumor: &message.RumorRequest{Contents: s}},
		response,
	)
	return *response
}

func addPeer(conn *net.UDPConn, r *http.Request) message.ValidationResponse {
	decoder := json.NewDecoder(r.Body)
	var s struct {
		Address string
		Port    string
	}
	err := decoder.Decode(&s)
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	response := &message.ValidationResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			AddPeer: &message.AddPeerRequest{Address: s.Address, Port: s.Port},
		},
		response,
	)
	return *response
}

func sendPrivateMessage(
	conn *net.UDPConn,
	r *http.Request,
) message.ValidationResponse {
	decoder := json.NewDecoder(r.Body)
	var s struct {
		Contents    string
		Destination string
	}
	err := decoder.Decode(&s)
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	response := &message.ValidationResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			SendPrivate: &message.PrivatePutRequest{
				Contents:    s.Contents,
				Destination: s.Destination,
			},
		},
		response,
	)
	return *response
}

func shareFile(
	conn *net.UDPConn,
	r *http.Request,
) message.FileShareResponse {
	file, header, err := r.FormFile("file")
	if err != nil {
		return message.FileShareResponse{Name: "Error: No File", Metakey: ""}
	}
	var Buf bytes.Buffer
	defer file.Close()
	io.Copy(&Buf, file)
	request, response := files.ShareFile(Buf, header.Filename)
	if err != nil || len(response.Metakey) == 0 {
		return *response
	}
	gossiperResponse := &message.FileShareResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			FileShare: &message.FileShareRequest{
				Name:     request.Name,
				Size:     request.Size,
				Metafile: request.Metafile,
				Metahash: request.Metahash,
			},
		},
		gossiperResponse,
	)
	return *gossiperResponse
}

func downloadFile(
	conn *net.UDPConn,
	r *http.Request,
) message.ValidationResponse {
	decoder := json.NewDecoder(r.Body)
	var s struct {
		Metakey  string
		Filename string
		Origin   string
	}
	err := decoder.Decode(&s)
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	response := &message.ValidationResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			Download: &message.FileDownloadRequest{
				Name:     s.Filename,
				Metahash: files.KeyToHash(s.Metakey),
				Origin:   &s.Origin,
			},
		},
		response,
	)
	return *response
}

func contactGossiper(
	conn *net.UDPConn,
	request *message.ClientPacket,
	response interface{},
) {
	buf, _ := protobuf.Encode(request)
	conn.Write(buf)
	for {
		buf := make([]byte, 1024000)
		_, _ = conn.Read(buf)
		protobuf.Decode(buf, response)
		return
	}
}

func main() {
	port := flag.String(
		"port",
		"8000",
		"port for the web handler (\"default 8000\")",
	)
	gossiper := flag.String(
		"gossiper",
		"8080",
		"port of the gossiper",
	)
	flag.Parse()

	gossiperPort = *gossiper
	http.HandleFunc("/", multiplexer)
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
