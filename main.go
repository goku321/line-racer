package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/goku321/line-racer/master"
	"github.com/goku321/line-racer/model"
	"github.com/goku321/line-racer/racer"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// flags
	nodeType := flag.String("nodeType", "master", "type of node: master/racer")
	masterIP := flag.String("masterIP", "127.0.0.1", "ip address of master process")
	racers := flag.Int("racers", 2, "number of racers")
	racerID := flag.Int("racerID", 0, "unique racer id (0 <= id < number of racers")
	laps := flag.Int("laps", 10, "number of lap")
	ip := flag.String("ip", "127.0.0.1", "ip address of the node")
	port := flag.String("port", "3000", "port to use")
	flag.Parse()

	_, err := strconv.ParseInt(*port, 10, 64)
	if err != nil {
		log.Fatalf("error parsing port number: %s", *port)
	}

	if *nodeType == "master" {
		m := master.New(*ip, *port, *racers, *laps)
		m.GenerateLaps()
		go m.Listen()
		m.WaitForRacers()
		m.StartRace()
		m.PrintLaps()
	} else {
		r := racer.New(*racerID, *ip, *port, *masterIP)
		pingMsg := model.NewPingMessage(r.Port+":"+strconv.Itoa(r.ID), r.Master+":3000")
		r.SignalMaster(pingMsg)
		r.ListenForNewLap()
	}
}
