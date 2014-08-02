package lib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type WatcherConfig struct {
	Dir, Ext       string
	Files, Plugins []string
}

type Watcher interface {
	GetAllFiles() int
	OutChan() chan *File
}

type watcher struct {
	Root, Dir string
	OutC      chan *File
}

func (w *watcher) OutChan() chan *File {
	return w.OutC
}

type FileWatcher struct {
	watcher
	Files []string
}

func (w *FileWatcher) GetAllFiles() int {
	size := 0
	for _, name := range w.Files {
		size++
		path := filepath.Join(w.Root, w.Dir, name)
		w.OutC <- getFile(path)
	}
	return size
}

type DirWatcher struct {
	watcher
	Ext        string
	Processors []*Processor
}

func (w *DirWatcher) GetAllFiles() int {
	size := 0
	dir := filepath.Join(w.Root, w.Dir)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, w.Ext) {
			size++
			f := getFile(path)

			for _, p := range w.Processors {
				p.InC <- f
			}
		}
		return err
	})
	return size
}

func NewWatcher(root string, outC chan *File, c *WatcherConfig, config *Config) Watcher {
	if len(c.Files) > 0 {
		return &FileWatcher{
			watcher: watcher{root, c.Dir, outC},
			Files:   c.Files,
		}
	}

	w := DirWatcher{
		watcher: watcher{root, c.Dir, outC},
		Ext:     c.Ext,
	}

	for _, name := range c.Plugins {
		p := config.GetProcessor(name)

		if p.PipeTo == "" {
			p.OutC = w.OutC
		}
		w.Processors = append(w.Processors, p)
	}
	return &w
}

func getFile(path string) *File {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return &File{Name: path, Content: string(b)}
}
