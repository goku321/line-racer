package master

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/goku321/line-racer/model"
)

// Master manages all the racers
type Master struct {
	IPAddr          string
	Port            string
	racersCount     int
	racers          []string
	laps            []lap
	currentLapCount int
	mutex           sync.Mutex
}

// lap represent a single lap
type lap struct {
	number      int
	pos         []model.Point
	start       string
	end         string
	timeElapsed int
}

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
	Source      string
	Dest        string
	Type        string
	Coordinates []model.Point
}

// NewLap generates a new lap
func NewLap(number int, pos []model.Point) *lap {
	return &lap{
		number: number,
		pos:    pos,
	}
}

// New inits new master
func New(ip, port string, racersCount int) *Master {
	return &Master{
		IPAddr:          ip,
		Port:            port,
		racers:          []string{},
		racersCount:     racersCount,
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
func (m *Master) Listen() {
	ln, err := net.Listen("tcp", ":"+m.Port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("master listening on %s:%s", m.IPAddr, m.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn, m)
	}
}

func handleConnection(conn net.Conn, m *Master) {
	defer conn.Close()
	s := conn.RemoteAddr().String()
	log.Printf("Serving %s\n", s)

	var msg Message
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Print(err)
	}

	if msg.Type == "ready" {
		// register racer
		m.registerRacer(msg.Source)

		// assign a unique index (0 <= index < N)
		m.mutex.Lock()
		id := len(m.racers) - 1
		m.mutex.Unlock()

		if err = json.NewEncoder(conn).Encode(&id); err != nil {
			log.Fatal(err)
		}
		go m.SendLap(msg.Source)

	} else if msg.Type == "update" {
		// do nothing
		log.Printf("racer %s position update: (%d, %d)", msg.Source, msg.Coordinates[0].X, msg.Coordinates[0].Y)
	}
}

// SendLap sends a lap to racers
func (m *Master) SendLap(racer string) {
	laddr, err := net.ResolveTCPAddr("tcp", "")
	if err != nil {
		log.Fatalf("error resolving tcp address: %v", err)
	}

	raddr, err := net.ResolveTCPAddr("tcp", racer)
	if err != nil {
		log.Fatalf("error resolving tcp address: %s, reason: %v", racer, err)
	}

	for {
		conn, err := net.DialTCP("tcp", laddr, raddr)
		if err != nil {
			log.Print("failed to establish connection to racer, retrying...", err)
			time.Sleep(time.Second * 5)
		} else {
			// Send current lap
			r := m.laps[m.currentLapCount]
			newMsg := getNewMessage(m.IPAddr+":"+m.Port, racer, r.pos)
			newMsg.Type = "race"
			err := json.NewEncoder(conn).Encode(&newMsg)
			if err != nil {
				log.Printf("error sending lap to racer %s", racer)
			}
			break
		}
	}
}

func getNewMessage(source, dest string, c []model.Point) Message {
	return Message{
		Source:      source,
		Dest:        dest,
		Type:        "ready",
		Coordinates: c,
	}
}

// GenerateLaps generates 10 laps for n racers
func (m *Master) GenerateLaps() {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	// 10 is total number of laps
	for i := 0; i < 10; i++ {
		l := []model.Point{}
		for j := 0; j < m.racersCount; j++ {
			p := model.New(r.Intn(50000), r.Intn(50000))
			l = append(l, p)
		}
		lap := NewLap(i, l)
		m.laps = append(m.laps, *lap)
	}
}

func (m *Master) registerRacer(r string) {
	m.mutex.Lock()
	m.racers = append(m.racers, r)
	m.mutex.Unlock()
}

// func calculateDistance() {
// 	for {
// 		if len(racers) >
// 	}
// }
