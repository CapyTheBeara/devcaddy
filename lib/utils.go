package lib

import (
	"fmt"
	"log"
)

var colors = map[string]string{
	"red":          "31",
	"error":        "31",
	"plugin error": "31",
	"watch error":  "31",
	"green":        "32",
	"watching":     "32",
	"yellow":       "33",
	"blue":         "34",
	"magenta":      "35",
	"server":       "35",
	"cyan":         "36",
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
