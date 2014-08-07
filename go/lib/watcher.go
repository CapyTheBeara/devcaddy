package lib

import (
	"log"
	"os"
	"path/filepath"
	"strings"
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
	Dir, Ext, Proxy string
	GroupAll        bool
	Files           []string
	PluginNames     []string `json:"plugins"`
}

type Watcher interface {
	GetAllFiles() int
	OutChan() chan *File
	Ready() chan bool
	PluginRes() chan *File
	addWatchDirs()
	matchesFile(string) bool
	fsWatcher() *fsnotify.Watcher
	listen(Watcher)
	handleNewDir(string)
	sendFileToPlugin(string, fsnotify.Op) int
}

func new_watcher(root, dir string, outC chan *File, c *WatcherConfig, config *Config) watcher {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	w := watcher{
		Root:    root,
		Dir:     dir,
		OutC:    outC,
		Proxy:   c.Proxy,
		fsw:     fsw,
		ready:   make(chan bool),
		procRes: make(chan *File),
		events:  make(map[string]time.Time),
		Plugins: []*Plugin{},
	}

	if len(c.PluginNames) == 0 {
		p := NewPlugin(&PluginConfig{}, func(f *File) *File {
			return f
		})
		p.OutC = w.procRes
		w.Plugins = append(w.Plugins, p)

	} else {
		for _, name := range c.PluginNames {
			p := config.GetPlugin(name)

			if p.PipeTo == "" {
				p.OutC = w.procRes
			}
			w.Plugins = append(w.Plugins, p)
		}
	}

	if c.GroupAll {
		w.store = NewStore(config.Root, &Config{})
	}
	return w
}

type watcher struct {
	Root, Dir, Proxy string
	OutC, procRes    chan *File
	ready            chan bool
	fsw              *fsnotify.Watcher
	events           map[string]time.Time
	store            *Store
	Plugins          []*Plugin
}

func (w *watcher) OutChan() chan *File {
	return w.OutC
}

func (w *watcher) Ready() chan bool {
	return w.ready
}

func (w *watcher) PluginRes() chan *File {
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

func (w *watcher) listen(wa Watcher) {
	// don't require reading from ready
	go func() {
		w.ready <- true
	}()

	go w.handlePluginRes()

	for {
		select {
		case evt := <-w.fsWatcher().Events:
			op := evt.Op
			now := time.Now()

			if evt.Op == fsnotify.Chmod {
				continue
			}

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

			w.sendFileToPlugin(evt.Name, op)

		case err := <-w.fsWatcher().Errors:
			if err != nil && err.Error() != "" {
				log.Println("watch error", err)
			}
		}
	}
}

func (w *watcher) sendFileToPlugin(opath string, op fsnotify.Op) int {
	path := opath

	if w.Proxy != "" {
		path = filepath.Join(w.Root, w.Proxy)
	}

	f := NewFile(path, op)

	if w.store != nil {
		w.store.Put(f.Name, f.Content)
		<-w.store.DidUpdate

		contents := []string{}
		for _, f := range w.store.GetAll() {
			contents = append(contents, f.Content)
		}
		f.Content = strings.Join(contents, "\n")
	}

	i := 0
	for _, p := range w.Plugins {
		i++
		p.InC <- f
	}

	return i
}

// TODO - don't need PluginRes anymore
func (w *watcher) handlePluginRes() {
	for {
		f := <-w.PluginRes()
		w.OutChan() <- f
	}
}
