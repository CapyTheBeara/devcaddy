package lib

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Root         string           `json:"root"`
	PluginConfs  []*PluginConfig  `json:"plugins"`
	WatcherConfs []*WatcherConfig `json:"watch"`
	Files        []*File          `json:"files"`
	Plugins      *Plugins         // TODO remove this
}

func NewConfig(cfg []byte) *Config {
	config := Config{}

	if len(cfg) == 0 {
		return &config
	}

	err := json.Unmarshal(cfg, &config)
	if err != nil {
		log.Fatalln("[error] Problem parsing JSON config:", err)
	}

	if config.Root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[error] Unable to determine your project folder name. Please try again or specify it in your .json config file.")
		}

		config.Root = "../" + filepath.Base(cwd)
	}

	for _, f := range config.Files {
		f.Type = "merge"
	}

	return &config
}

func (c *Config) GetPlugin(name string) *Plugin {
	return c.Plugins.Get(name)
}
