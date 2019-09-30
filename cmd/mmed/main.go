package main

import (
	"fmt"
	"runtime"

	log "github.com/coreswitch/log"

	"github.com/coreswitch/coreswitch/pkg/mme"
)

func main() {
	numCPUs := runtime.NumCPU()
	fmt.Println("numCPUs", numCPUs)
	log.SourceField = false
	log.FuncField = false

	server := mme.NewServer()
	server.Start()
	for {
	}
}
