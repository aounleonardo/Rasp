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
	"io/ioutil"
	"os"
	"math"
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
) message.ValidationResponse {
	var Buf bytes.Buffer
	file, header, err := r.FormFile("file")
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	defer file.Close()
	io.Copy(&Buf, file)
	request, err := bufferToShareRequest(Buf, header.Filename)
	if err != nil {
		return message.ValidationResponse{Success: false}
	}
	Buf.Reset()
	response := &message.ValidationResponse{}
	contactGossiper(conn, &message.ClientPacket{FileShare: request}, response)
	return *response
}

func bufferToShareRequest(
	buffer bytes.Buffer,
	filename string,
) (*message.FileShareRequest, error) {
	bufferLength := buffer.Len()
	bufferBytes := buffer.Bytes()
	nbChunks :=
		int(math.Ceil(float64(bufferLength) / float64(files.MaxFileChunkSize)))
	if nbChunks > files.MaxChunks {
		return nil, errors.New("file too big")
	}
	chunks := make(map[string][]byte)
	var metafile bytes.Buffer
	for chunk := 0; chunk < nbChunks; chunk++ {
		readChunk := make([]byte, files.MaxFileChunkSize)
		nbBytesRead, _ := buffer.Read(readChunk)
		hash := files.HashChunk(readChunk[:nbBytesRead])
		_, err := metafile.Write(hash)
		if err != nil {
			return nil, errors.New("error saving file: " + err.Error())
		}
		chunks[files.HashToKey(hash)] = readChunk[:nbBytesRead]
	}
	err := ioutil.WriteFile(files.SharedFiles+filename, bufferBytes, os.ModePerm)
	if err != nil {
		fmt.Println("error saving file", err.Error())
		return nil, err
	}
	metahash := files.HashChunk(metafile.Bytes())
	chunks[files.HashToKey(metahash)] = metafile.Bytes()
	for chunkName, chunk := range chunks {
		err = ioutil.WriteFile(
			files.Downloads+chunkName,
			chunk,
			os.ModePerm,
		)
		if err != nil {
			fmt.Println("error saving file", err.Error())
			return nil, err
		}
	}
	return &message.FileShareRequest{
		Name:     filename,
		Size:     uint32(bufferLength),
		Metafile: metafile.Bytes(),
		Metahash: metahash,
	}, nil
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
				Origin:   s.Origin,
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
