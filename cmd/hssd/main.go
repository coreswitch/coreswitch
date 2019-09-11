package main

import (
	"github.com/coreswitch/coreswitch/pkg/hss"
)

func main() {
	server := hss.NewServer()
	server.Start()
	for {
	}
}
