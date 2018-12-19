package main

import (
	"flag"
	"github.com/aounleonardo/Peerster/internal/pkg/gossip"
	"math/rand"
	"strings"
	"time"
)

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
		"249498",
		"name of the gossiper",
	)
	var peers = flag.String(
		"peers",
		"",
		"comma separated list of peers of the form ip:port",
	)
	var rtimer = flag.Int(
		"rtimer",
		0,
		"route rumors sending period in seconds, 0 to disable sending "+
			"of route rumors (default 0)",
	)
	var simple = flag.Bool(
		"simple",
		false,
		"run gossiper in simple broadcast mode",
	)
	flag.Parse()

	var peerList []string
	if len(*peers) > 0 {
		peerList = strings.Split(*peers, ",")
	}

	rand.Seed(time.Now().Unix())

	gossiper := gossip.NewGossiper(
		*name,
		*uiPort,
		*gossipAddr,
		peerList,
		*simple,
		*rtimer,
	)
	defer gossiper.ShutUp()

	gossiper.ListenForClientMessages()
}
