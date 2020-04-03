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
	racers          map[int]string
	posUpdates      []pos
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

// pos ...
type pos struct {
	id string
	model.Point
}

// Node represents a process
type Node struct {
	ID     string `json:"id"`
	IPAddr string `json:"ip_addr"`
	Port   string `json:"port"`
	Type   string `json:"type"`
	Status string `json:"status"`
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
		racers:          map[int]string{},
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

	var msg model.Message
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Print(err)
	}

	if msg.Type == "ready" {
		// assign a unique index (0 <= index < N)
		m.mutex.Lock()
		id := len(m.racers)
		m.mutex.Unlock()

		// register racer
		m.registerRacer(id, msg.Source)
		if err = json.NewEncoder(conn).Encode(&id); err != nil {
			log.Fatal(err)
		}
		go m.SendLap(msg.Source)

	} else if msg.Type == "pos" {
		log.Printf("racer %s position update: (%d, %d)", msg.Source, msg.Coordinates[0].X, msg.Coordinates[0].Y)
		m.updatePOS(msg.Source, msg.Coordinates[0])
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
			newMsg := model.NewMessage(m.IPAddr+":"+m.Port, racer, r.pos)
			newMsg.Type = "race"
			err := json.NewEncoder(conn).Encode(&newMsg)
			if err != nil {
				log.Printf("error sending lap to racer %s", racer)
			}
			break
		}
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

func (m *Master) registerRacer(id int, r string) {
	m.mutex.Lock()
	m.racers[id] = r
	m.mutex.Unlock()
}

func (m *Master) updatePOS(id string, p model.Point) {
	u := &pos{
		id:    id,
		Point: p,
	}
	m.mutex.Lock()
	m.posUpdates = append(m.posUpdates, *u)
	m.mutex.Unlock()
}

// CalculateDistance constantly polls a slice
func (m *Master) CalculateDistance() {
	for {
		if len(m.posUpdates) >= 2 {
			p1 := m.posUpdates[0]
			p2 := m.posUpdates[1]
			if p1.id == p2.id {
				m.mutex.Lock()
				m.posUpdates = m.posUpdates[1:]
				m.mutex.Unlock()
			} else {

				d := p1.Point.Distance(p2.Point)

				if d > 10 {
					// start a new lap
					log.Fatal("distance exceeds 10 units")
				}
			}
		}
	}
}
