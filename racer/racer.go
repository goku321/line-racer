package racer

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/goku321/line-racer/master"
	"github.com/goku321/line-racer/model"
)

// SignalMaster sends a signal to master process
// with its coordinates
func SignalMaster(n *master.Node, m *master.Message) {
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
			time.Sleep(time.Second * 5)
		} else {
			m.Type = "ready"
			m.Dest = master.NewNode("127.0.0.1", "3000", "master")
			err := json.NewEncoder(conn).Encode(&m)
			if err != nil {
				log.Fatalf("error communicating to master: %v", err)
			}
			var id int
			err = json.NewDecoder(conn).Decode(&id)
			log.Printf("id received from master %d", id)
			conn.Close()
			break
		}
	}
}

// ListenForNewCoordinates waits for master to get new coordinates
func ListenForNewCoordinates(n *master.Node) {
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
	log.Printf("new lap from %s\n", conn.RemoteAddr().String())

	var msg master.Message
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Print(err)
	}

	if msg.Type == "ready" {
		race(msg.Coordinates)
	}
}

func race(c []model.Point) {
	log.Printf("racing on lap %v", c)
}
