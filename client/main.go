package main

import (
	"flag"
	"runtime"
)

var flagConnect = flag.String("connect", "", "")

func init() {
	runtime.LockOSThread()
}

func main() {
	flag.Parse()
	game := NewGame()
	if *flagConnect != "" {
		game.Connect(*flagConnect)
	} else {
		game.MainMenu()
	}
}
