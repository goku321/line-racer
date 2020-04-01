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

// Message is a contract between master and racers
type Message struct {
	Source      Node
	Dest        Node
	Type        string
	Coordinates []int
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

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	s := conn.RemoteAddr().String()
	log.Printf("Serving %s\n", s)

	var msg Message
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Print(err)
	}

	if msg.Type == "ready" {
		// newMsg := getNewMessage(msg.Dest, msg.Source)
		log.Printf("%s is ready", s)
	}
}

// ConnectToRacer connects running node to n
func ConnectToRacer(n *Node, m *Message) {
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
			r := getNewMessage(*n, Node{IPAddr: "127.0.0.1", Port: "3000", Type: "master"})
			err := json.NewEncoder(conn).Encode(&r)
			if err != nil {
				log.Printf("error communicating to master: %v", err)
			}
			log.Printf("received from master: %v", r)
			break
		}
	}
}

func getNewMessage(source Node, dest Node) Message {
	return Message{
		Source:      source,
		Dest:        dest,
		Type:        "ready",
		Coordinates: []int{1, 2},
	}
}
