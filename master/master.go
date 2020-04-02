package master

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

var racers int
var mutex sync.Mutex

// Master manages all the racers
type Master struct {
	racers          []string
	laps            []lap
	currentLapCount int
	mutex           sync.Mutex
}

// lap represent a single lap
type lap struct {
	number      int
	pos         [][]int
	start       string
	end         string
	timeElapsed int
}

var laps map[int]lap

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

// New inits new master
func New() *Master {
	return &Master{
		racers:          []string{},
		laps:            []lap{},
		currentLapCount: 0,
		mutex:           sync.Mutex{},
	}
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
		mutex.Lock()
		racerIndex := racers
		racers++
		mutex.Unlock()
		err := json.NewEncoder(conn).Encode(&racerIndex)
		if err != nil {
			log.Printf("error communicating with %s", msg.Source.ID)
		}
		go ConnectToRacer(&msg.Dest, &msg.Source, nil)
	} else if msg.Type == "update" {
		// do nothing
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
		Coordinates: generateNewLap(2),
	}
}

func generateNewLap(racersCount int) [][]int {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	lap := [][]int{}

	for i := 0; i < racersCount; i++ {
		lap = append(lap, []int{r.Intn(50000), r.Intn(50000)})
	}

	return lap
}

// func calculateDistance() {
// 	for {
// 		if len(racers) >
// 	}
// }
