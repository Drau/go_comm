package main

import "log"

type logger struct {
	color string
}

func (l logger) print(m ...string) {
	message := ""
	for _, v := range m {
		message += v
	}
	log.Print(l.color + message + "\033[0m")
}
