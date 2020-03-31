package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// Node represents a process
type Node struct {
	ID     string `json:"id"`
	IPAddr string `json:"ip_addr"`
	Port   string `json:"port"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type message struct {
	Source      Node
	Dest        Node
	Type        string
	coordinates []int
}

// NewNode inits and returns new node
func NewNode(ip, port, t string) *Node {
	return &Node{
		ID:     fmt.Sprintf("%s:%s", ip, port),
		IPAddr: ip,
		Port:   port,
		Type:   t,
		Status: "down",
	}
}

// Listen starts listening on a port
func Listen(n *Node) {
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

		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Serving %s\n", conn.RemoteAddr().String())

	var msg message
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Print(err)
	}

	log.Print(msg)
}

// ConnectToNode connects running node to n
func ConnectToNode(n *Node) {
	laddr, err := net.ResolveTCPAddr("tcp", n.ID)
	if err != nil {
		log.Fatalf("error resolving tcp address: %s, reason: %v", n.ID, err)
	}

	raddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatalf("error resolving tcp address: %v", err)
	}

	for {
		_, err := net.DialTCP("tcp", laddr, raddr)
		if err != nil {
			log.Print("failed to establish connection to master, retrying...")
		}
		log.Print("waiting for master node to be available")
	}
}
