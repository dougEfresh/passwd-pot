package queue

import "github.com/dougEfresh/passwd-pot/api"

//EventQueue sends events somewhere
type EventQueue interface {
	Send(event *api.Event)
}
