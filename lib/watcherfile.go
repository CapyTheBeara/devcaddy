package lib

import (
	"path/filepath"
)

type FileWatcher struct {
	watcher
	Files []string
}

func (w *FileWatcher) GetAllFiles() int {
	size := 0
	for _, name := range w.Files {
		size++
		path := filepath.Join(w.Root, w.Dir, name)
		w.OutC <- w.getFile(path)
	}
	return size
}

func (w *FileWatcher) matchesFile(name string) bool {
	for _, f := range w.Files {
		if name == filepath.Join(w.Dir, f) {
			return true
		}
	}
	return false
}

func (w *FileWatcher) addWatchDirs() {
	for _, f := range w.Files {
		path := filepath.Dir(filepath.Join(w.Root, w.Dir, f))
		w.addWatchDir(path)
	}
}

func (w *FileWatcher) handleNewDir(name string) {
	return
}

func (w *FileWatcher) processFile(f *File) int {
	w.OutC <- f
	return 1
}
