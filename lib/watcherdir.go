package lib

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type DirWatcher struct {
	watcher
	Ext        string
	Processors []*Processor
}

func (w *DirWatcher) GetAllFiles() int {
	size := 0
	dir := filepath.Join(w.Root, w.Dir)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalln("[error] Problem getting files:", err)
		}
		if !info.IsDir() && strings.HasSuffix(path, w.Ext) {
			f := w.getFile(path)
			size += w.processFile(f)
		}
		return nil
	})
	return size
}

func (w *DirWatcher) matchesFile(name string) bool {
	return strings.HasPrefix(name, w.fullPath()) && strings.HasSuffix(name, w.Ext)
}

func (w *DirWatcher) addWatchDirs() {
	filepath.Walk(w.fullPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalln("[error] Couldn't watch folder:", err)
		}

		if info.IsDir() {
			log.Printf("[watching] %s for .%s files", path, w.Ext)
			w.addWatchDir(path)
		}
		return nil
	})
}

func (w *DirWatcher) handleNewDir(name string) {
	w.addWatchDir(name)
}

func (w *DirWatcher) processFile(f *File) int {
	i := 0
	for _, p := range w.Processors {
		if !p.LogOnly && !p.NoOutput {
			i++
		}
		p.InC <- f
	}

	return i
}

func (w *DirWatcher) fullPath() string {
	return filepath.Join(w.Root, w.Dir)
}
