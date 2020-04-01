package app

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

var racers []string

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
	Coordinates [][]int
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
		newMsg := getNewMessage(msg.Dest, msg.Source)
		log.Printf("%s is ready", s)
		go ConnectToRacer(&msg.Dest, &msg.Source, &newMsg)
	} else if msg.Type == "running" {
		// calculate distance
	}
}

// ConnectToRacer connects running node to n
func ConnectToRacer(master, racer *Node, m *Message) {
	laddr, err := net.ResolveTCPAddr("tcp", master.ID)
	if err != nil {
		log.Fatalf("error resolving tcp address: %s, reason: %v", master.ID, err)
	}

	raddr, err := net.ResolveTCPAddr("tcp", racer.ID)
	if err != nil {
		log.Fatalf("error resolving tcp address: %s, reason: %v", racer.ID, err)
	}

	for {
		conn, err := net.DialTCP("tcp", laddr, raddr)
		if err != nil {
			log.Print("failed to establish connection to racer, retrying...")
			time.Sleep(time.Second * 5)
		} else {
			r := getNewMessage(*master, *racer)
			err := json.NewEncoder(conn).Encode(&r)
			if err != nil {
				log.Printf("error communicating to racer: %v", err)
			}
			conn.Close()
			break
		}
	}
}

func getNewMessage(source Node, dest Node) Message {
	return Message{
		Source:      source,
		Dest:        dest,
		Type:        "ready",
		Coordinates: generateNewLap(),
	}
}

func generateNewLap() [][]int {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	return [][]int{{r.Intn(50000), r.Intn(50000)}, {r.Intn(50000), r.Intn(50000)}}
}
