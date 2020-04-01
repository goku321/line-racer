package racer

import (
	"encoding/json"
	"log"
	"net"

	"github.com/goku321/line-racer/app"
)

// SignalMaster sends a signal to master process
// with its coordinates
func SignalMaster(n *app.Node, m *app.Message) {
	laddr, err := net.ResolveTCPAddr("tcp", n.ID)
	if err != nil {
		log.Fatalf("error resolving tcp address: %s, reason: %v", n.ID, err)
	}

	raddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatalf("error resolving tcp address: %v", err)
	}

	for {
		conn, err := net.DialTCP("tcp", laddr, raddr)
		if err != nil {
			log.Print("failed to establish connection to master, retrying...")
		} else {
			err := json.NewEncoder(conn).Encode(&m)
			if err != nil {
				log.Fatalf("error communicating to master: %v", err)
			}
			conn.Close()
			break
		}
	}
}

// ListenForNewCoordinates waits for master to get new coordinates
func ListenForNewCoordinates(n *app.Node) {
	ln, err := net.Listen("tcp", ":"+n.Port)
	if err != nil {
		log.Fatal(err)
	}

	n.Status = "up"
	log.Printf("%s listening on %s:%s", n.Type, n.IPAddr, n.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Serving %s\n", conn.RemoteAddr().String())

	var msg app.Message
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Print(err)
	}

	if msg.Type == "new" {
		race(msg.Coordinates)
	}
}

func race(c []int) {
	log.Print("racing")
}
