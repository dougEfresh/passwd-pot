package work

import (
	"github.com/dougEfresh/passwd-pot/cmd/queue"
	"sync"
)

//Worker holds context for all pot holders
type Worker struct {
	Addr       string
	EventQueue queue.EventQueue
	Wg         sync.WaitGroup
}
