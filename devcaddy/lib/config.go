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
		log.Fatalln(err, ERROR_CONFIG_PARSE)
	}

	if config.Root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln(ERROR_CONFIG_ROOT)
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
