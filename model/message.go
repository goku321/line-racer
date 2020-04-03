package model

// Message is a contract between master and racers
type Message struct {
	Source      string
	Dest        string
	Type        string
	Coordinates []Point
}

// NewMessage returns a new message
func NewMessage(source, dest string, c []Point) Message {
	return Message{
		Source:      source,
		Dest:        dest,
		Type:        "ready",
		Coordinates: c,
	}
}
