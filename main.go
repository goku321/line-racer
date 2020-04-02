package main

import (
	"flag"
	"log"
	"strconv"

	master "github.com/goku321/line-racer/master"
	"github.com/goku321/line-racer/racer"
)

func main() {
	nodeType := flag.String("nodeType", "racer", "type of node: master/racer")
	racers := flag.Int("racers", 2, "number of racers")
	clusterIP := flag.String("clusterIP", "127.0.0.1", "ip address of the node")
	port := flag.String("port", "3003", "port to use")
	flag.Parse()

	_, err := strconv.ParseInt(*port, 10, 64)
	if err != nil {
		log.Fatalf("error parsing port number: %s", *port)
	}

	n := master.NewNode(*clusterIP, *port, *nodeType)

	if n.Type == "master" {
		m := master.New(*racers)
		m.GenerateLaps()
		master.Listen(n, m)
	}

	racer.SignalMaster(n, &master.Message{Source: n})
 	racer.ListenForNewCoordinates(n)
}
