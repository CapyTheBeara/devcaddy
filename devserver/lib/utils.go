package lib

import (
	"fmt"
	"log"
)

var colors = map[string]string{
	"red":      "31",
	"error":    "31",
	"info":     "31",
	"green":    "32",
	"created":  "32",
	"watching": "32",
	"yellow":   "33",
	"blue":     "34",
	"magenta":  "35",
	"removed":  "35",
	"server":   "35",
	"cyan":     "36",
	"modified": "36",
}

func Color(kind, text string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", colors[kind], text)
}

var Plog logging = false

type logging bool

func (l *logging) Printf(format string, args ...interface{}) {
	if Plog {
		log.Printf(format, args...)
	}
}

func (l *logging) Println(kind string, args ...interface{}) {
	if Plog {
		log.Println(args...)
	}
}

func (l *logging) PrintC(kind, text string) {
	if Plog {
		log.Println(Color(kind, "["+kind+"] "+text))
	}
}

func LogProcessedFiles(in chan *lib.File, done chan bool, size int) chan *lib.File {
	out := make(chan *lib.File)
	go func() {
		i := 0
		init := true
		for {
			f := <-in

			switch f.Op {
			case lib.LOG:
				if f.Content != "" {
					lib.Plog.PrintC("info", f.Content)
				}
			case lib.CREATE:
				if !init {
					lib.Plog.PrintC("created", f.Name)
				}
			case lib.WRITE:
				if !init {
					lib.Plog.PrintC("modified", f.Name)
				}
			case lib.REMOVE:
				lib.Plog.PrintC("removed", f.Name)
			case lib.RENAME:
				lib.Plog.PrintC("removed", f.Name)
			case lib.ERROR:
				lib.Plog.PrintC("error", f.Error.Error())
				lib.Plog.PrintC("error", f.Content)
			}

			if init {
				i++
				if i >= size {
					done <- true
					init = false
				}
			}
			out <- f
		}
	}()

	return out
}
