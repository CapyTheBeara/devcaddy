package lib

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/fsnotify.v0"
)

func NewWatcher(root string, outC chan *File, c *WatcherConfig, config *Config) Watcher {
	var wa Watcher
	ready := make(chan bool)

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	if len(c.Files) > 0 {
		w := &FileWatcher{
			watcher: watcher{root, c.Dir, outC, ready, fsw},
			Files:   c.Files,
		}
		wa = w
	} else {
		w := DirWatcher{
			watcher: watcher{root, c.Dir, outC, ready, fsw},
			Ext:     c.Ext,
		}

		for _, name := range c.Plugins {
			p := config.GetProcessor(name)

			if p.PipeTo == "" {
				p.OutC = w.OutC
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

func (w *watcher) listen(wa Watcher) {
	w.ready <- true
	for {
		select {
		case evt := <-w.fsWatcher().Events:
			op := evt.Op

			if evt.Name != "" {
				log.Println(evt)
			}

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
				Name: filepath.Join(w.Root, evt.Name),
				Op:   op,
			}

			if op != fsnotify.Remove {
				f = *w.getFile(f.Name)
			}

			wa.processFile(&f)
			// w.OutC <- &f

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
