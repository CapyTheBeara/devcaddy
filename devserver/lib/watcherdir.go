package lib

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type DirWatcher struct {
	watcher
	name string
	Ext  string
}

func (w *DirWatcher) Name() string {
	if w.name != "" {
		return w.name
	}
	return w.Dir + ":" + w.Ext
}

func (w *DirWatcher) GetAllFiles() int {
	size := 0
	dir := filepath.Join(w.Root, w.Dir)
	skip := false

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if skip {
			return filepath.SkipDir
		}

		if err != nil {
			log.Fatalln("[error] Problem getting files:", err)
		}

		if !info.IsDir() && strings.HasSuffix(path, w.Ext) {
			size += w.sendFileToPlugin(path, CREATE)

			if w.Proxy != "" {
				skip = true
			}
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
			Plog.PrintC("watching", "*."+w.Ext+": "+path)
			w.addWatchDir(path)
		}
		return nil
	})
}

func (w *DirWatcher) handleNewDir(name string) {
	w.addWatchDir(name)
}

func (w *DirWatcher) fullPath() string {
	return filepath.Join(w.Root, w.Dir)
}
