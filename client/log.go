package main

import (
	"log"
)

func padRight(str, pad string) string {
	str += pad
	return str
}

func padLeft(str, pad string) string {
	str = pad + str
	return str
}

func fixedLen(str string, leng int) string {
	for i := 0; len(str) < leng; i++ {
		if i%2 == 0 {
			str = padLeft(str, " ")
		} else {
			str = padRight(str, " ")
		}
	}
	return str
}

func Log(component string, lvl string, msg string) {
	component = fixedLen(component, 8)
	lvl = fixedLen(lvl, 6)
	log.Printf("[%s] [%s] - %s\n", component, lvl, msg)
}
