package racer

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/goku321/line-racer/model"
)

// Racer represents a racer
type Racer struct {
	ID     int
	IPAddr string
	Port   string
	Master string
	Laps   [][]model.Point
	Status string
}

// New returns a new racer type
func New(id int, ip, port, masterIP string) *Racer {
	return &Racer{
		ID:     id,
		IPAddr: ip,
		Port:   port,
		Status: "up",
		Master: masterIP,
	}
}

// NewConnection returns new tcp connection
func (r *Racer) NewConnection() (*net.TCPConn, error) {
	srcAddr, err := net.ResolveTCPAddr("tcp", "")
	if err != nil {
		return nil, err
	}

	dstAddr, err := net.ResolveTCPAddr("tcp", r.Master+":3000")
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", srcAddr, dstAddr)
}

// SignalMaster sends a signal to master process
// with its coordinates
func (r *Racer) SignalMaster(m *model.Message) {
	for {
		conn, err := r.NewConnection()
		if err != nil {
			log.Printf("connecting to master, %v", err)
			time.Sleep(time.Second * 5)
		} else {
			if err = json.NewEncoder(conn).Encode(&m); err != nil {
				log.Fatalf("error communicating to master: %v", err)
			}

			if err = conn.Close(); err != nil {
				log.Fatal("unable to close connection")
			}
			break
		}
	}
}

// SendPOSUpdate sends position updates to master every 50ms
func (r *Racer) SendPOSUpdate(m *model.Message) {
	for {
		conn, err := r.NewConnection()
		if err != nil {
			log.Printf("racer %d: connecting to master, %v", r.ID, err)
			conn.Close()
			time.Sleep(time.Second * 5)
		} else {
			if err = json.NewEncoder(conn).Encode(&m); err != nil {
				log.Printf("racer %d: error communicating to master: %v", r.ID, err)
			}
			conn.Close()
			break
		}
	}
}

// ListenForNewLap waits for master to get new coordinates
func (r *Racer) ListenForNewLap() {
	ln, err := net.Listen("tcp", ":"+r.Port)
	if err != nil {
		log.Fatalf("racer %d: %v", r.ID, err)
	}

	log.Printf("racer %d: listening on %s:%s", r.ID, r.IPAddr, r.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn, r)
	}
}

func handleConnection(conn net.Conn, r *Racer) {
	log.Printf("racer %d: new lap from master", r.ID)

	var msg model.Message
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Printf("racer %d: %v", r.ID, err)
	}

	// close connection here as message has already been received
	conn.Close()

	if msg.Type == "race" {
		r.Laps = append(r.Laps, msg.Coordinates)
		r.race(msg.Coordinates)
	} else if msg.Type == "kill" {
		log.Printf("racer %d: kill signal received. racer will terminate", r.ID)
		r.printLaps()
		os.Exit(0)
	}
}

func (r *Racer) race(l []model.Point) {
	log.Printf("racer %d: racing on lap %v", r.ID, l)
	// add a check for invalid lap
	m, c := l[r.ID].X, l[r.ID].Y
	p := getStartingPoint(l)
	log.Printf("racer %d: starting race from (%d, %d)", r.ID, p.X, p.Y)

	for {
		time.Sleep(time.Millisecond * 50)
		p.X++
		p.Y = (m * p.X) + c
		m := &model.Message{
			Source:      strconv.Itoa(r.ID),
			Dest:        r.Master + ":3000",
			Type:        "pos",
			Coordinates: []model.Point{p},
		}
		r.SendPOSUpdate(m)
	}
}

func (r *Racer) printLaps() {
	for k, v := range r.Laps {
		log.Printf("racer %d: lap %d: %v", r.ID, k+1, v)
	}
}

func getStartingPoint(x []model.Point) model.Point {
	m1, c1, m2, c2 := x[0].X, x[0].Y, x[1].X, x[1].Y

	sX := (c1 - c2) / (m2 - m1)
	sY := ((m2 * c1) - (m1 * c2)) / (m2 - m1)

	return model.New(sX, sY)
}
