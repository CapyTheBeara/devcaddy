package lib

import (
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/fsnotify.v0"
)

var (
	zeroTime = time.Time{}
	_events  = Events{contents: make(map[string]time.Time)}
)

type Events struct {
	sync.Mutex
	contents map[string]time.Time
}

func (es *Events) Set(e *Event) {
	es.Lock()
	es.contents[e.Id] = e.CreatedAt
	es.Unlock()
}

func (es *Events) Get(evt *Event) time.Time {
	return es.contents[evt.Id]
}

func (es *Events) ShouldIgnore(e *Event) bool {
	prevTime := _events.Get(e)
	return prevTime != zeroTime && e.CreatedAt.Sub(prevTime).Seconds() < 0.01
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

func (e *Event) Ignore() bool {
	if e.Op == CHMOD {
		return true
	}

	ignore := _events.ShouldIgnore(e)
	_events.Set(e)
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
