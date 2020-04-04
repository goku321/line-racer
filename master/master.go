package master

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/goku321/line-racer/model"
)

var wg sync.WaitGroup

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

	} else if msg.Type == "pos" {
		log.Printf("racer %s position update: (%d, %d)", msg.Source, msg.Coordinates[0].X, msg.Coordinates[0].Y)
		m.updatePOS(msg.Source, msg.Coordinates[0])
	}
}

// SendLap sends a lap to racers
func (m *Master) SendLap(racer string, msg model.Message) {
	defer wg.Done()
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
			// Send Lap
			err := json.NewEncoder(conn).Encode(&msg)
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
				m.mutex.Lock()
				m.posUpdates = m.posUpdates[2:]
				m.mutex.Unlock()
				d := p1.Point.Distance(p2.Point)

				if d > 10 {
					// start a new lap
					log.Print("distance exceeds 10 units")
					break
				}
			}
		}
	}
}

// WaitForRacers waits infinitely for racers to get connected
func (m *Master) WaitForRacers() {
	for {
		if len(m.racers) == m.racersCount {
			break
		}
	}
}

// StartRace inits race
func (m *Master) StartRace() {
	for _, v := range m.laps {
		wg.Add(m.racersCount)
		for _, r := range m.racers {
			lapMsg := model.NewMessage(m.IPAddr+":"+m.Port, r, v.pos)
			lapMsg.Type = "race"
			go m.SendLap(r, lapMsg)
		}
		wg.Wait()
		m.CalculateDistance()
	}
	m.SendKillMessage()
}

// SendKillMessage sends a kill message to all racers
func (m *Master) SendKillMessage() {
	wg.Add(m.racersCount)
	for _, r := range m.racers {
		msg := model.NewMessage(m.IPAddr+":"+m.Port, r, []model.Point{})
		msg.Type = "kill"
		go m.SendLap(r, msg)
	}
	wg.Wait()
}
