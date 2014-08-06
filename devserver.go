package main

import (
	"io/ioutil"
	"log"
	"strconv"

	"github.com/monocle/devserver/lib"
)

func main() {
	log.SetFlags(log.Ltime)
	lib.Plog = true

	done := make(chan bool)

	cfg, err := ioutil.ReadFile("devserver_config.json")
	if err != nil {
		log.Fatalln("[error] Problem reading devserver_config.json", err)
	}

	c := lib.NewConfig(cfg)
	store := c.PopulateStore(done)

	<-done

	lib.Plog.PrintC("server", strconv.Itoa(len(store.Files))+" files defined in store")
	lib.StartServer(store, "4200", "", "assets")

}
