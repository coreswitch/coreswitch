package main

import (
	"fmt"
	"runtime"

	"github.com/coreswitch/coreswitch/pkg/mme"
)

func main() {
	numCPUs := runtime.NumCPU()
	fmt.Println("numCPUs", numCPUs)

	server := mme.NewServer()
	server.Start()
	for {
	}
}
