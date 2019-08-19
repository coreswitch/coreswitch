package mme

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/coreswitch/coreswitch/pkg/s1ap"
	"github.com/ishidawataru/sctp"
)

type ServerConfig struct {
	retryTime time.Duration
}

// Server is MME top level structure.
type Server struct {
	conf ServerConfig
	ln   *sctp.SCTPListener
	wg   sync.WaitGroup
	done chan interface{}
}

func NewServer() *Server {
	return &Server{
		conf: ServerConfig{
			retryTime: 30,
		},
	}
}

func SCTPInfoSize() int {
	info := sctp.SndRcvInfo{}
	return int(unsafe.Sizeof(info))
}

func (s *Server) serveClient(conn net.Conn, infoSize int) error {
	for {
		bufsize := 2048
		buf := make([]byte, bufsize+128) // add overhead of SCTPSndRcvInfoWrappedConn

		n, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if n < infoSize {
			return fmt.Errorf("n (%d) < SCTPinfoSize (%d)", n, infoSize)
		}
		log.Printf("Read length: %d", n)

		buf = buf[infoSize:n]

		fmt.Printf("len of buf %d\n", len(buf))
		for i := 0; i < len(buf); i++ {
			fmt.Printf("%2x ", buf[i])
		}
		fmt.Printf("\n")

		s1ap.Decode(buf)
	}
}

func (s *Server) sctpListen() (*sctp.SCTPListener, error) {
	ips := []net.IPAddr{}

	addr := &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    S1AP_PORT_NUMBER,
	}

	return sctp.ListenSCTP("sctp", addr)
}

// Start() function initiate MME services.
func (s *Server) Start() error {
	if s.done != nil {
		return fmt.Errorf("Server already started")
	}
	s.wg.Add(1)
	s.done = make(chan interface{})
	infoSize := SCTPInfoSize()

	go func() {
		defer s.wg.Done()
		for {
		retry:
			ln, err := s.sctpListen()
			s.ln = ln
			if err != nil {
				fmt.Println(err.Error())
				select {
				case <-s.done:
					return
				case <-time.After(s.conf.retryTime * time.Second):
					goto retry
				}
			}

			log.Printf("Listen on %s\n", ln.Addr())

			for {
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println(err.Error())
					select {
					case <-s.done:
						return
					default:
						goto retry
					}
				}
				log.Printf("Accepted Connection from RemoteAddr: %s", conn.RemoteAddr())

				wconn := sctp.NewSCTPSndRcvInfoWrappedConn(conn.(*sctp.SCTPConn))

				go s.serveClient(wconn, infoSize)
			}
		}
	}()

	return nil
}

// Stop() will block until all of the goroutines stop.
func (s *Server) Stop() error {
	if s.done == nil {
		return fmt.Errorf("Server already stopped")
	}
	close(s.done)
	s.wg.Wait()
	s.done = nil

	return nil
}
