package main

import (
	"fmt"
	"runtime"

	"github.com/coreswitch/coreswitch/pkg/mme"
)

type S1 struct {
	i1 int `asn1: ""`
	i2 int `asn1: ""`
}

func main() {
	numCPUs := runtime.NumCPU()
	fmt.Println("numCPUs", numCPUs)

	server := mme.NewServer()
	server.Start()
	for {
	}
}
