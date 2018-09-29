package main

import (
	"flag"
	"strings"
	"net"
)

type Gossiper struct {
	Name    string
	conn    *net.UDPConn
	address *net.UDPAddr
}

func NewGossiper(
	name,
	uiPort,
	gossipAddr string,
	peer_list []string,
	simple bool,
) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", gossipAddr)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	println("Creating New Gossiper")
	return &Gossiper{
		Name:    name,
		conn:    udpConn,
		address: udpAddr,
	}
}

func main() {
	var uiPort = flag.String(
		"UIPort",
		"8080",
		"port for the UI client (default \"8080\")",
	)
	var gossipAddr = flag.String(
		"gossipAddr",
		"127.0.0.1:5000",
		"ip:port for the gossiper (default \"127.0.0.1:5000\")",
	)
	var name = flag.String(
		"name",
		"Leo",
		"name of the gossiper",
	)
	var peers = flag.String(
		"peers",
		"",
		"comma separated list of peers of the form ip:port",
	)
	var simple = flag.Bool(
		"simple",
		false,
		"run gossiper in simple broadcast mode",
	)
	flag.Parse()

	var peer_list []string
	if len(*peers) > 0 {
		peer_list = strings.Split(*peers, ",")
	}

	var _ = NewGossiper(*name, *uiPort, *gossipAddr, peer_list, *simple)
}
