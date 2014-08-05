package lib

import (
	"log"
)

var Plog logging = false

type logging bool

func (l *logging) Printf(format string, args ...interface{}) {
	if Plog {
		log.Printf(format, args...)
	}
}

func (l *logging) Println(args ...interface{}) {
	if Plog {
		log.Println(args...)
	}
}
