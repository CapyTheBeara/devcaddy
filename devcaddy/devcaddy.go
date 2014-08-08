package main

import (
	"io/ioutil"
	"log"
	"strconv"

	"github.com/monocle/devcaddy/devcaddy/lib"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	lib.Plog = true

	done := make(chan bool)
	watcherOutput := make(chan *lib.File)

	cfg, err := ioutil.ReadFile("devcaddy.json")
	if err != nil {
		log.Fatalln("[error] Problem reading devcaddy.json", err)
	}
	c := lib.NewConfig(cfg)

	plugins := lib.NewPlugins(c.PluginConfs)
	// TODO remove this
	c.Plugins = plugins

	watchers := lib.NewWatchers(c, watcherOutput)
	size := watchers.GetInitialFiles()

	store := lib.NewStore(c)
	store.Input = lib.LogProcessedFiles(watcherOutput, done, size)
	store.Listen()

	<-done
	lib.Plog.PrintC("server", strconv.Itoa(len(store.Files))+" files defined in store")
	lib.StartServer(store, "4200", "", "assets")
}
