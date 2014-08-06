package main

import (
	"io/ioutil"
	"log"
	"strconv"

	"github.com/monocle/devcaddy/lib"
)

func main() {
	log.SetFlags(log.Ltime)
	lib.Plog = true

	done := make(chan bool)

	cfg, err := ioutil.ReadFile("devcaddy_config.json")
	if err != nil {
		log.Fatalln("[error] Problem reading devcaddy_config.json", err)
	}

	c := lib.NewConfig(cfg)
	store := c.PopulateStore(done)

	<-done

	lib.Plog.PrintC("server", strconv.Itoa(len(store.Files))+" files defined in store")
	lib.StartServer(store, "4200", "", "assets")

}
