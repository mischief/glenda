package core

type MessageType int

const (
	Message MessageType = iota
	Connected
)

// Connector connects to a chat network.
type Connector interface {
	Nick() string
	Run() error
	Quit(msg string)
	Message(who, msg string)
	Handle(what MessageType, h EventHandler)
}
