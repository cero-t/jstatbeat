package main

import (
	jsatbeat "./beat"
	"github.com/elastic/libbeat/beat"
)

func main() {
	beat.Run("jstatbeat", "0.1", jsatbeat.New())
}
