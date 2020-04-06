package master

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goku321/line-racer/model"
)

var wg sync.WaitGroup

// Master manages all the racers
type Master struct {
	IPAddr      string
	Port        string
	racersCount int
	racers      map[string]string
	posUpdates  []pos // can be a queue
	posRecords  []pos
	laps        []lap
	lapsCount   int
	racerMutex  sync.Mutex
	posMutex    sync.Mutex
}

// lap represent a single lap
type lap struct {
	number      int
	pos         []model.Point
	start       time.Time
	end         time.Time
	timeElapsed int64
	latencies   map[string]int64
}

// pos represents  position of a racer
type pos struct {
	id         string
	updateTime time.Time
	model.Point
}

// NewLap returns a new lap
func NewLap(number int, pos []model.Point) *lap {
	return &lap{
		number: number,
		pos:    pos,
	}
}

// New returns new master
func New(ip, port string, racersCount, lapsCount int) *Master {
	return &Master{
		IPAddr:      ip,
		Port:        port,
		racers:      map[string]string{},
		racersCount: racersCount,
		laps:        []lap{},
		lapsCount:   lapsCount,
	}
}

// NewConnection returns new tcp connection
func (m *Master) NewConnection(racer string) (*net.TCPConn, error) {
	srcAddr, err := net.ResolveTCPAddr("tcp", "")
	if err != nil {
		return nil, err
	}

	dstAddr, err := net.ResolveTCPAddr("tcp", racer)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", srcAddr, dstAddr)
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

	if msg.Type == "ping" {
		s := strings.Split(msg.Source, ":")
		port := s[0]
		id, err := strconv.Atoi(s[1])

		if err != nil {
			log.Printf("invalid racer id %v from %v", id, port)
		}

		if id < 0 || id >= m.racersCount {
			log.Fatal("invalid racer id")
		}

		addr := strings.Split(conn.RemoteAddr().String(), ":")
		// register racer
		m.registerRacer(s[1], addr[0]+":"+port)

		log.Printf("racer %d connected - %v", id, time.Now())

	} else if msg.Type == "pos" {
		log.Printf("racer %s position update: (%d, %d)", msg.Source, msg.Coordinates[0].X, msg.Coordinates[0].Y)
		m.updatePOS(msg.Source, msg.Coordinates[0])
	}
}

// SendMessage sends a lap to racers
func (m *Master) SendMessage(racer string, msg *model.Message) {
	defer wg.Done()

	for {
		conn, err := m.NewConnection(racer)
		if err != nil {
			log.Print("master: failed to establish connection to racer, retrying...", err)
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

	for i := 0; i < m.lapsCount; i++ {
		l := []model.Point{}
		for j := 0; j < m.racersCount; j++ {
			p := model.New(r.Intn(50000), r.Intn(50000))
			l = append(l, p)
		}
		lap := NewLap(i, l)
		m.laps = append(m.laps, *lap)
	}
}

func (m *Master) registerRacer(id string, r string) {
	m.racers[id] = r
}

func (m *Master) updatePOS(id string, p model.Point) {
	u := &pos{
		id:         id,
		updateTime: time.Now(),
		Point:      p,
	}
	m.posMutex.Lock()
	m.posUpdates = append(m.posUpdates, *u)
	m.posRecords = append(m.posRecords, *u)
	m.posMutex.Unlock()
}

// CalculateDistance constantly polls a slice
func (m *Master) CalculateDistance() {
	for {
		if len(m.posUpdates) >= 2 {
			p1 := m.posUpdates[0]
			p2 := m.posUpdates[1]
			if p1.id == p2.id {
				m.posMutex.Lock()
				m.posUpdates = m.posUpdates[1:]
				m.posMutex.Unlock()
			} else {
				m.posMutex.Lock()
				m.posUpdates = m.posUpdates[2:]
				m.posMutex.Unlock()
				d := p1.Point.Distance(p2.Point)

				if d > 10 {
					// start a new lap
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
	for k, v := range m.laps {
		wg.Add(m.racersCount)

		start := time.Now()
		for _, r := range m.racers {
			lapMsg := model.NewMessage(m.IPAddr+":"+m.Port, r, "race", v.pos)
			lapMsg.Type = "race"
			go m.SendMessage(r, lapMsg)
		}
		wg.Wait()
		m.CalculateDistance()
		end := time.Now()

		latencies := m.calculateLatency(start)
		m.updateLap(k, start, end, latencies)
		// Clear position update queue
		m.posUpdates = []pos{}
	}
	m.SendKillMessage()
}

// SendKillMessage sends a kill message to all racers
func (m *Master) SendKillMessage() {
	wg.Add(m.racersCount)
	for _, r := range m.racers {
		msg := model.NewMessage(m.IPAddr+":"+m.Port, r, "kill", []model.Point{})
		go m.SendMessage(r, msg)
	}
	wg.Wait()
}

// PrintLaps prints all the laps
func (m *Master) PrintLaps() {
	sort.Slice(m.laps, func(i, j int) bool {
		return m.laps[i].timeElapsed < m.laps[j].timeElapsed
	})
	for k, v := range m.laps {
		log.Printf("Lap %d %v %s %s %d %d %d", k+1, v.pos, v.start.Format("03:04:05"), v.end.Format("03:04:05"), v.timeElapsed, v.latencies["0"], v.latencies["1"])
	}
}

// updateLap updates start and end time for a lap
func (m *Master) updateLap(index int, start, end time.Time, latencies map[string]int64) {
	l := m.laps[index]
	l.start = start
	l.end = end
	l.timeElapsed = end.Sub(start).Milliseconds()
	l.latencies = latencies
	m.laps[index] = l
}

// calculateLatency calculates latencies for each racer
func (m *Master) calculateLatency(lapStart time.Time) map[string]int64 {
	latencies := map[string]int64{}
	for r := range m.racers {
		racerPos := []pos{}
		// copy racer's position updates to a new slice
		for _, u := range m.posRecords {
			if r == u.id {
				racerPos = append(racerPos, u)
			}
		}

		// sort the slice
		sort.Slice(racerPos, func(i, j int) bool {
			return racerPos[i].X < racerPos[j].X
		})

		// add difference b/w the time first update was received
		// lap start time
		lat := 50 - (racerPos[0].updateTime.Sub(lapStart))
		// calculate latency
		for i := 1; i < len(racerPos); i++ {
			t1, t2 := racerPos[i-1], racerPos[i]
			lat += 50 - (t2.updateTime.Sub(t1.updateTime))
		}

		var totalUpdates int64
		totalUpdates = int64(len(racerPos) + 1)

		latencies[r] = lat.Milliseconds() / totalUpdates
	}
	return latencies
}
