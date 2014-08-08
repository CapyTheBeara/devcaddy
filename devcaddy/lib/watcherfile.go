package lib

import (
	"path/filepath"
)

type FileWatcher struct {
	watcher
	name  string
	Files []string
}

func (w *FileWatcher) Name() string {
	return w.name
}

// TODO - check for proxy
func (w *FileWatcher) GetAllFiles() int {
	size := 0
	for _, name := range w.Files {
		size++
		path := filepath.Join(w.Root, w.Dir, name)
		w.sendFileToPlugin(NewPseudoEvent(path, CREATE))
	}
	return size
}

func (w *FileWatcher) IsWatchingEvent(evt *Event) bool {
	for _, f := range w.Files {
		if evt.Name() == filepath.Join(w.Root, w.Dir, f) {
			return true
		}
	}
	return false
}

func (w *FileWatcher) addWatchDirs() {
	for _, f := range w.Files {
		name := filepath.Join(w.Root, w.Dir, f)

		Plog.PrintC("watching", name)
		w.addWatchDir(filepath.Dir(name))
	}
}

func (w *FileWatcher) handleNewDir(e *Event) {
	return
}
