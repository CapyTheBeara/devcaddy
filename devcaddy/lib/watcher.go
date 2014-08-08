package lib

import (
	"log"
	"path/filepath"

	"gopkg.in/fsnotify.v0"
)

func NewWatcher(root string, out chan *File, c *WatcherConfig, config *Config) Watcher {
	var wa Watcher
	_watcher := new_watcher(root, c.Dir, out, c, config)

	if len(c.Files) > 0 {
		wa = &FileWatcher{
			watcher: _watcher,
			name:    c.Name,
			Files:   c.Files,
		}

	} else {
		wa = &DirWatcher{
			watcher: _watcher,
			name:    c.Name,
			Ext:     c.Ext,
		}
	}

	go _watcher.listen(wa)
	go _watcher.sendReady()
	wa.addWatchDirs()
	return wa
}

type WatcherConfig struct {
	Dir, Ext    string
	Name, Proxy string
	GroupAll    bool
	Files       []string
	PluginNames []string `json:"plugins"`
}

type Watcher interface {
	Name() string
	GetAllFiles() int
	Ready() chan bool
	IsWatchingEvent(*Event) bool
	addWatchDirs()
	fsWatcher() *fsnotify.Watcher
	handleNewDir(*Event)
	sendFileToPlugin(*Event) int
}

func (w *watcher) Store() *Store {
	return w.store
}

func new_watcher(root, dir string, out chan *File, c *WatcherConfig, config *Config) watcher {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	w := watcher{
		Root:    root,
		Dir:     dir,
		Proxy:   c.Proxy,
		fsw:     fsw,
		ready:   make(chan bool),
		Plugins: NewPlugins([]*PluginConfig{}),
	}

	if len(c.PluginNames) == 0 {
		c.PluginNames = append(c.PluginNames, "_identity_")
	}

	for _, name := range c.PluginNames {
		p := config.GetPlugin(name)
		p.SetOutC(out)
		w.Plugins.Add(p)
	}

	if c.GroupAll {
		w.store = NewStore(&Config{Root: config.Root})
	}
	return w
}

type watcher struct {
	Root       string
	Dir, Proxy string
	ready      chan bool
	fsw        *fsnotify.Watcher
	store      *Store
	Plugins    *Plugins
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

func (w *watcher) sendReady() {
	w.ready <- true
}

func (w *watcher) listen(wa Watcher) {
	for {
		select {
		case evt := <-w.fsWatcher().Events:
			e := NewEvent(evt, wa)

			if e.Ignore() {
				continue
			}

			if !wa.IsWatchingEvent(e) {
				// ?need to handle dir deletion?
				if e.IsNewDir() {
					wa.handleNewDir(e)
				}
				continue
			}

			w.sendFileToPlugin(e)

		case err := <-w.fsWatcher().Errors:
			w.sendFileToPlugin(NewPseudoEvent("watcher error", ERROR, err))
		}
	}
}

func (w *watcher) sendFileToPlugin(e *Event) int {
	if w.Proxy != "" {
		e.Event.Name = filepath.Join(w.Root, w.Proxy)
	}
	f := NewFile(e)

	if w.store != nil {
		w.store.PutFile(f)
		<-w.store.DidUpdate

		f = NewFile(e)
		f.Content = w.store.GetAllContents()
	}

	return w.Plugins.Each(func(p *Plugin) {
		p.InC <- f
	})
}

func NewWatchers(c *Config, out chan *File) *Watchers {
	content := map[string]Watcher{}

	for _, f := range c.Files {
		wc := &WatcherConfig{
			Name:        f.Name,
			Dir:         f.Dir,
			Ext:         f.Ext,
			Files:       f.Files,
			PluginNames: f.PluginNames,
		}
		w := NewWatcher(c.Root, out, wc, c)
		content[w.Name()] = w
	}

	for _, wc := range c.WatcherConfs {
		w := NewWatcher(c.Root, out, wc, c)
		content[w.Name()] = w
	}

	return &Watchers{content}
}

type Watchers struct {
	content map[string]Watcher
}

func (ws *Watchers) Get(name string) Watcher {
	w := ws.content[name]
	return w
}

func (ws *Watchers) GetInitialFiles() int {
	size := 0
	for _, w := range ws.content {
		size += w.GetAllFiles()
	}
	return size
}
