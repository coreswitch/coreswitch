package hss

import (
	"fmt"

	"github.com/fiorix/go-diameter/diam"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() {
	fmt.Println("Start HSS server")
	diam.ListenAndServeNetwork("tcp", ":3868", handler, nil)
}

func (s *Server) Stop() {
	fmt.Println("Stop HSS server")
}
