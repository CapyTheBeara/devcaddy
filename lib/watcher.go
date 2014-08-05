package lib

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/fsnotify.v0"
)

var zeroTime = time.Time{}

func NewWatcher(root string, outC chan *File, c *WatcherConfig, config *Config) Watcher {
	var wa Watcher
	ready := make(chan bool)
	procRes := make(chan *File)
	events := make(map[string]time.Time)

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	_watcher := watcher{root, c.Dir, outC, ready, fsw, procRes, events}

	if len(c.Files) > 0 {
		w := &FileWatcher{
			watcher: _watcher,
			Files:   c.Files,
		}
		wa = w
	} else {
		w := DirWatcher{
			watcher: _watcher,
			Ext:     c.Ext,
		}

		for _, name := range c.Plugins {
			p := config.GetProcessor(name)

			if p.PipeTo == "" {
				p.OutC = w.procRes
			}
			w.Processors = append(w.Processors, p)
		}
		wa = &w
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
	processFile(*File) int
}

type watcher struct {
	Root, Dir string
	OutC      chan *File
	ready     chan bool
	fsw       *fsnotify.Watcher
	procRes   chan *File
	events    map[string]time.Time
}

func (w *watcher) OutChan() chan *File {
	return w.OutC
}

func (w *watcher) Ready() chan bool {
	return w.ready
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

func (w *watcher) ProcessorRes() chan *File {
	return w.procRes
}

func (w *watcher) filterProcessorRes() {
	for {
		f := <-w.ProcessorRes()
		if f == nil || f.LogOnly {
			continue
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

			// if evt.Name != "" {
			// 	log.Println(evt)
			// }

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

			f := File{
				Name: evt.Name,
				Op:   op,
			}

			if op != fsnotify.Remove && op != fsnotify.Rename {
				f = *w.getFile(f.Name)
			}

			wa.processFile(&f)

		case err := <-w.fsWatcher().Errors:
			if err != nil && err.Error() != "" {
				log.Println("[watch error]", err)
			}
		}
	}
}

func (w *watcher) getFile(path string) *File {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	return &File{Name: path, Content: string(b)}
}
