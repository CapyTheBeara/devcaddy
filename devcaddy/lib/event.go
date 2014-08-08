package lib

import (
	"log"
	"os"
	"time"

	"gopkg.in/fsnotify.v0"
)

var (
	zeroTime = time.Time{}
	// events   = map[string]time.Time{}
)

func init() {
	// events = make(map[string]time.Time)
}

func NewEvent(e fsnotify.Event, w Watcher) *Event {
	return &Event{
		Event:     e,
		Id:        w.Name() + e.Name,
		name:      w.Name(),
		Op:        FileOp(e.Op),
		CreatedAt: time.Now(),
	}
}

func NewPseudoEvent(name string, op FileOp, args ...error) *Event {
	var err error
	if args != nil {
		err = args[0]
	}

	return &Event{
		name:      name,
		Op:        op,
		Error:     err,
		CreatedAt: time.Now(),
	}
}

type Event struct {
	fsnotify.Event
	Id        string
	name      string
	Op        FileOp
	Error     error
	CreatedAt time.Time
}

func (e *Event) Name() string {
	if e.Event.Name != "" {
		return e.Event.Name
	}
	return e.name
}

// TODO remove param
func (e *Event) Ignore(events map[string]time.Time) bool {
	if e.Op == CHMOD {
		return true
	}

	ignore := false

	if events[e.Id] != zeroTime && e.CreatedAt.Sub(events[e.Id]).Seconds() < 0.01 {
		ignore = true
	}

	events[e.Id] = e.CreatedAt
	return ignore
}

func (e *Event) IsNewDir() bool {
	if e.Op == CREATE {
		fi, err := os.Stat(e.Event.Name)
		if err != nil {
			log.Println("[error] Unable to get file info:", err)
		}

		if fi.IsDir() {
			return true
		}
	}
	return false
}
