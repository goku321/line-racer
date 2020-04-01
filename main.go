package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/goku321/line-racer/app"
	"github.com/goku321/line-racer/racer"
)

func main() {
	nodeType := flag.String("nodeType", "racer", "type of node: master/racer")
	clusterIP := flag.String("clusterIP", "127.0.0.1", "ip address of the node")
	port := flag.String("port", "3000", "port to use")
	flag.Parse()

	_, err := strconv.ParseInt(*port, 10, 64)
	if err != nil {
		log.Fatalf("error parsing port number: %s", *port)
	}

	n := app.NewNode(*clusterIP, *port, *nodeType)

	if n.Type == "master" {
		app.Listen(n)
	}

	racer.SignalMaster(n, nil)
	racer.ListenForNewCoordinates(n)
}
