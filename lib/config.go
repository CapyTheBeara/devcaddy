package lib

import (
	"encoding/json"
	"log"
)

type Config struct {
	Root       string
	Watch      []*WatcherConfig
	Watchers   []Watcher
	Plugins    []*ProcessorConfig
	Processors []*Processor
	Files      []*File
	Store      *Store
	outC       chan *File
}

func (c *Config) GetProcessor(name string) *Processor {
	for _, p := range c.Processors {
		if name == p.Name {
			return p
		}
	}

	log.Fatalln("Plugin not defined:", name)
	return nil
}

// TODO move this to where it makes more sense
func (c *Config) PopulateStore(done chan bool) *Store {
	size := -1
	sendUpdate := false

	go func() {
		i := 0
		for f := range c.outC {
			i++

			if f.IsDeleted() {
				c.Store.Delete(f.Name)
			} else {
				c.Store.Put(f.Name, f.Content, sendUpdate)
			}

			if size != -1 && i == size {
				sendUpdate = true
				done <- true
			}
		}
	}()

	c.makeWatchers()

	n := 0
	for _, w := range c.Watchers {
		n += w.GetAllFiles()
	}

	size = n
	return c.Store
}

func (c *Config) makeProcessors() {
	for _, pl := range c.Plugins {
		p := NewCommandProcessor(pl)

		if pl.PipeTo != "" {
			pipeP := c.GetProcessor(pl.PipeTo)
			p.OutC = pipeP.InC
		}

		c.Processors = append(c.Processors, p)
	}
	return
}

func (c *Config) makeWatchers() {
	c.outC = make(chan *File)

	for _, f := range c.Files {
		wc := WatcherConfig{
			Dir:     f.Dir,
			Ext:     f.Ext,
			Files:   f.Files,
			Plugins: f.Plugins,
		}
		c.Watch = append(c.Watch, &wc)
	}

	for _, wc := range c.Watch {
		w := NewWatcher(c.Root, c.outC, wc, c)
		c.Watchers = append(c.Watchers, w)
	}
}

func NewConfig(cfg []byte) *Config {
	config := Config{}
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		log.Fatalln("[error] Problem parsing JSON config:", err)
	}

	for _, f := range config.Files {
		f.Type = "merge"
	}

	config.makeProcessors()
	config.Store = NewStore(config.Root, &config)

	return &config
}
