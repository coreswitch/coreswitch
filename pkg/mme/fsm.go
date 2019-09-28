package mme

import "fmt"

// FSM for MME session.
type FSM struct {
}

func (f *FSM) hello() {
	fmt.Println("hello")
}
