package model

// Message is a contract between master and racers
type Message struct {
	Source      string
	Dest        string
	Type        string
	Coordinates []Point
}

// NewMessage returns a new message
func NewMessage(source, dest, msgType string, c []Point) *Message {
	return &Message{
		Source:      source,
		Dest:        dest,
		Type:        msgType,
		Coordinates: c,
	}
}

// NewPingMessage returns a new message
func NewPingMessage(source, dest string) *Message {
	return &Message{
		Source: source,
		Dest:   dest,
		Type:   "ping",
	}
}
