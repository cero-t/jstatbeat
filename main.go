package main

import (
	jsatbeat "./beat"
	"github.com/elastic/libbeat/beat"
)

func main() {
	jsb := &jsatbeat.Jstatbeat{}
	b := beat.NewBeat("jstatbeat", "0.1", jsb)
	b.CommandLineSetup()
	b.LoadConfig()
	jsb.Config(b)
	b.Run()
}
