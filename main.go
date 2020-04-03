package main

import (
	"flag"
	"log"
	"strconv"
	"sync"

	master "github.com/goku321/line-racer/master"
	"github.com/goku321/line-racer/racer"
)

var wg sync.WaitGroup

func main() {
	nodeType := flag.String("nodeType", "master", "type of node: master/racer")
	racers := flag.Int("racers", 2, "number of racers")
	clusterIP := flag.String("clusterIP", "127.0.0.1", "ip address of the node")
	port := flag.String("port", "3000", "port to use")
	flag.Parse()

	_, err := strconv.ParseInt(*port, 10, 64)
	if err != nil {
		log.Fatalf("error parsing port number: %s", *port)
	}

	n := master.NewNode(*clusterIP, *port, *nodeType)

	if n.Type == "master" {
		m := master.New(*clusterIP, *port, *racers)
		m.GenerateLaps()
		go m.CalculateDistance()
		m.Listen()
	}

	r := racer.New(*clusterIP, *port)
	r.SignalMaster(&master.Message{Source: r.IPAddr + ":" + r.Port})
	r.ListenForNewCoordinates()
}
