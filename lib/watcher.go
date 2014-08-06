package lib

import (
	"log"
	"os"
	"time"

	"gopkg.in/fsnotify.v0"
)

var zeroTime = time.Time{}

func NewWatcher(root string, outC chan *File, c *WatcherConfig, config *Config) Watcher {
	var wa Watcher
	_watcher := new_watcher(root, c.Dir, outC, c, config)

	if len(c.Files) > 0 {
		wa = &FileWatcher{
			watcher: _watcher,
			Files:   c.Files,
		}

	} else {
		wa = &DirWatcher{
			watcher: _watcher,
			Ext:     c.Ext,
		}
	}

	go wa.listen(wa)
	wa.addWatchDirs()
	return wa
}

type WatcherConfig struct {
	Dir, Ext       string
	Files, Plugins []string
}

type Watcher interface {
	GetAllFiles() int
	OutChan() chan *File
	Ready() chan bool
	ProcessorRes() chan *File
	addWatchDirs()
	matchesFile(string) bool
	fsWatcher() *fsnotify.Watcher
	listen(Watcher)
	handleNewDir(string)
	processFile(string, fsnotify.Op) int
}

func new_watcher(root, dir string, outC chan *File, c *WatcherConfig, config *Config) watcher {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	w := watcher{
		Root:       root,
		Dir:        dir,
		OutC:       outC,
		fsw:        fsw,
		ready:      make(chan bool),
		procRes:    make(chan *File),
		events:     make(map[string]time.Time),
		Processors: []*Processor{},
	}

	if len(c.Plugins) == 0 {
		p := NewProcessor(&ProcessorConfig{}, func(f *File) *File {
			return f
		})
		p.OutC = w.procRes
		w.Processors = append(w.Processors, p)

	} else {
		for _, name := range c.Plugins {
			p := config.GetProcessor(name)

			if p.PipeTo == "" {
				p.OutC = w.procRes
			}
			w.Processors = append(w.Processors, p)
		}
	}
	return w
}

type watcher struct {
	Root, Dir     string
	OutC, procRes chan *File
	ready         chan bool
	fsw           *fsnotify.Watcher
	events        map[string]time.Time
	Processors    []*Processor
}

func (w *watcher) OutChan() chan *File {
	return w.OutC
}

func (w *watcher) Ready() chan bool {
	return w.ready
}

func (w *watcher) ProcessorRes() chan *File {
	return w.procRes
}

func (w *watcher) fsWatcher() *fsnotify.Watcher {
	return w.fsw
}

func (w *watcher) addWatchDir(path string) {
	err := w.fsw.Add(path)
	if err != nil {
		log.Fatal(err)
	}
}

func (w *watcher) filterProcessorRes() {
	for {
		f := <-w.ProcessorRes()

		if f == nil || f.LogOnly {
			continue
		}

		if f.Error != nil {
			Plog.PrintC("plugin error", f.Content)
		}
		w.OutChan() <- f
	}
}

func (w *watcher) listen(wa Watcher) {
	// don't require reading from ready
	go func() {
		w.ready <- true
	}()

	go w.filterProcessorRes()

	for {
		select {
		case evt := <-w.fsWatcher().Events:
			op := evt.Op
			now := time.Now()

			if w.events[evt.Name] != zeroTime {
				if now.Sub(w.events[evt.Name]).Seconds() < 0.01 {
					continue
				}
			}

			w.events[evt.Name] = now

			if !wa.matchesFile(evt.Name) {
				// ?need to handle dir deletion?
				if op == fsnotify.Create {
					fi, err := os.Stat(evt.Name)
					if err != nil {
						log.Println("[error] Unable to get file info:", err)
					}

					if fi.IsDir() {
						wa.handleNewDir(evt.Name)
					}
				}
				continue
			}

			wa.processFile(evt.Name, op)

		case err := <-w.fsWatcher().Errors:
			if err != nil && err.Error() != "" {
				log.Println("watch error", err)
			}
		}
	}
}

func (w *watcher) processFile(path string, op fsnotify.Op) int {
	f := NewFile(path, op)

	i := 0
	for _, p := range w.Processors {
		if !p.LogOnly && !p.NoOutput {
			i++
		}
		p.InC <- f
	}

	return i
}
