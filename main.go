package main

import (
	"flag"
	"log"
	"strconv"

	master "github.com/goku321/line-racer/master"
	"github.com/goku321/line-racer/racer"
)

func main() {
	nodeType := flag.String("nodeType", "master", "type of node: master/racer")
	clusterIP := flag.String("clusterIP", "127.0.0.1", "ip address of the node")
	port := flag.String("port", "3000", "port to use")
	flag.Parse()

	_, err := strconv.ParseInt(*port, 10, 64)
	if err != nil {
		log.Fatalf("error parsing port number: %s", *port)
	}

	n := master.NewNode(*clusterIP, *port, *nodeType)

	if n.Type == "master" {
		master.Listen(n)
	}

	racer.SignalMaster(n, &master.Message{Source: *n})
	racer.ListenForNewCoordinates(n)
}
