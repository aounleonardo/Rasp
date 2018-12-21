package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/aounleonardo/Peerster/internal/pkg/chain"
	"github.com/aounleonardo/Peerster/internal/pkg/files"
	"github.com/aounleonardo/Peerster/internal/pkg/message"
	"github.com/dedis/protobuf"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
		regexp.MatchString("/chats/", r.RequestURI); isChatRequest {
		return json.Marshal(waitForChats(conn))
	}
	if isPrivateMessageRequest, _ :=
		regexp.MatchString("/pm/*/*/*/", r.RequestURI); isPrivateMessageRequest {
		partner, unordered, ordered := getPrivateIndexes(r.RequestURI)
		return json.Marshal(waitForPrivates(conn, partner, unordered, ordered))
	}
	if isSearchRequest, _ :=
		regexp.MatchString("/searches/", r.RequestURI); isSearchRequest {
		return json.Marshal(waitForSearches(conn))
	}

	if isPlayersRequest, _ :=
		regexp.MatchString("/players/", r.RequestURI); isPlayersRequest {
			return json.Marshal(waitForPlayers(conn))

	}

	if isStateRequest, _ :=
		regexp.MatchString("/state/", r.RequestURI); isStateRequest {
			return json.Marshal(waitForStates(conn))
	}

	return nil, errors.New("unsupported URI")
}

func postHandler(r *http.Request, conn *net.UDPConn) ([]byte, error) {
	if isMessagesRequest, _ :=
		regexp.MatchString("/message/", r.RequestURI); isMessagesRequest {
		return json.Marshal(readMessage(conn, r))
	}
	if isPeerRequest, _ := regexp.MatchString("/peers/", r.RequestURI); isPeerRequest {
		return json.Marshal(addPeer(conn, r))
	}
	if isPrivateMessageRequest, _ :=
		regexp.MatchString("/pm/", r.RequestURI); isPrivateMessageRequest {
		return json.Marshal(sendPrivateMessage(conn, r))
	}
	if isFileShareRequest, _ :=
		regexp.MatchString("/share-file/", r.RequestURI); isFileShareRequest {
		return json.Marshal(shareFile(conn, r))
	}
	if isDownloadRequest, _ :=
		regexp.MatchString("/download-file/", r.RequestURI); isDownloadRequest {
		fmt.Println("isDownloadRequest")
		return json.Marshal(downloadFile(conn, r))
	}
	if isSearchRequest, _ :=
		regexp.MatchString("/search-for/*/", r.RequestURI); isSearchRequest {
		return json.Marshal(
			searchForKeywords(conn, getSearchKeywords(r.RequestURI)))
	}

	if isCreateMatchRequest, _ :=
		regexp.MatchString("/create-match/", r.RequestURI); isCreateMatchRequest {
		return json.Marshal(sendMatchRequest(conn, r))
	}

	if isRespondMatchRequest, _ :=
		regexp.MatchString("/accept-match/", r.RequestURI); isRespondMatchRequest {
		return json.Marshal(sendMatchResponse(conn, r))

	}
	return nil, errors.New("unsupported URI")
}

func waitForPlayers(conn *net.UDPConn) *chain.PlayersResponse {
	response := &chain.PlayersResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{Players: &chain.PlayersRequest{}},
		response)

	return response
}

func waitForStates(conn *net.UDPConn) *chain.StateResponse {
	response := &chain.StateResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{States: &chain.StateRequest{}},
		response)

	return response
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
		Origin   *string
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
				Origin:   s.Origin,
			},
		},
		response,
	)
	return *response
}

func searchForKeywords(
	conn *net.UDPConn,
	keywords []string,
) message.ValidationResponse {
	response := &message.ValidationResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			Search: &message.PerformSearchRequest{
				Keywords: keywords,
				Budget:   nil,
			},
		},
		response,
	)
	return *response
}

func getSearchKeywords(uri string) []string {
	keywords := strings.TrimSuffix(
		strings.TrimPrefix(uri, "/search-for/"),
		"/",
	)
	return strings.Split(keywords, ",")
}

func waitForSearches(conn *net.UDPConn) message.SearchesResponse {
	response := &message.SearchesResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{GetSearches: &message.SearchesRequest{}},
		response,
	)
	return *response
}

func sendMatchRequest(
	conn *net.UDPConn,
	r *http.Request,
) message.ValidationResponse {
	decoder := json.NewDecoder(r.Body)
	var req struct {
		Destination *string
		Bet         chain.Bet
		Move        chain.Move
	}
	err := decoder.Decode(&req)
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	response := &message.ValidationResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			CreateMatch: &chain.CreateMatchRequest{
				Destination: req.Destination,
				Bet:         req.Bet,
				Move:        req.Move,
			},
		},
		response,
	)
	return *response
}

func sendMatchResponse(
	conn *net.UDPConn,
	r *http.Request,
) message.ValidationResponse {
	decoder := json.NewDecoder(r.Body)
	var res struct {
		Identifier chain.Uid
		Move       chain.Move
	}
	err := decoder.Decode(&res)
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	response := &message.ValidationResponse{}
	contactGossiper(
		conn,
		&message.ClientPacket{
			AcceptMatch: &chain.AcceptMatchRequest{
				Identifier: res.Identifier,
				Move:       res.Move,
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
		"port for the web server (\"default 8000\")",
	)
	gossiper := flag.String(
		"gossiper",
		"8080",
		"port of the gossiper",
	)
	flag.Parse()

	gossiperPort = *gossiper
	http.HandleFunc("/", multiplexer)
	fmt.Printf("Listening on %s for gossiper %s", *port, *gossiper)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
