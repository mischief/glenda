package core

import (
	"golang.org/x/net/context"
)

type ResponseWriter interface {
	Message(who, what string)
}

type EventType interface {
	noimpl()
}

type EventMessage struct {
	Type MessageType
}

func (EventMessage) noimpl() {}

type EventCommand struct {
	Command string
}

func (EventCommand) noimpl() {}

type EventHandler interface {
	Handle(context.Context, ResponseWriter, *Event) error
}

type EventHandlerFunc func(context.Context, ResponseWriter, *Event) error

func (f EventHandlerFunc) Handle(ctx context.Context, rw ResponseWriter, e *Event) error {
	return f(ctx, rw, e)
}

type Event struct {
	// orignal sender
	Sender string
	// target of message, chan for public and nick for private
	Target string
	// message
	Args string
}
