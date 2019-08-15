package mme

import (
	"bytes"
	"encoding/asn1"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/fiorix/go-diameter/diam/dict"
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

var helloDictionary = xml.Header + `
<diameter>
        <application id="12" type="acct">
                <command code="11" short="HM" name="Hello-Message">
                        <request>
                                <rule avp="Session-Id" required="true" max="1"/>
                                <rule avp="Origin-Host" required="true" max="1"/>
                                <rule avp="Origin-Realm" required="true" max="1"/>
                                <rule avp="User-Name" required="false" max="1"/>
                        </request>
                        <answer>
                                <rule avp="Session-Id" required="true" max="1"/>
                                <rule avp="Result-Code" required="true" max="1"/>
                                <rule avp="Origin-Host" required="true" max="1"/>
                                <rule avp="Origin-Realm" required="true" max="1"/>
                                <rule avp="Error-Message" required="false" max="1"/>
                        </answer>
                </command>
        </application>
</diameter>
`

func (s *Server) serveClient(conn net.Conn) error {
	for {
		// diameter.
		dict.Default.Load(bytes.NewReader([]byte(helloDictionary)))

		// m, err := diam.ReadMessage(conn, dict.Default)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// } else {
		// 	log.Println(m)
		// }

		bufsize := 2048
		buf := make([]byte, bufsize+128) // add overhead of SCTPSndRcvInfoWrappedConn

		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("read failed: %v", err)
			return err
		}
		log.Printf("read: %d", n)

		var sndRcvInfoSize uintptr

		info := sctp.SndRcvInfo{}
		sndRcvInfoSize = unsafe.Sizeof(info)
		fmt.Println("RcvInfoSize", int(sndRcvInfoSize))

		buf = buf[:n]
		n = n - int(sndRcvInfoSize)
		buf = buf[int(sndRcvInfoSize):]
		fmt.Printf("len of buf %d\n", len(buf))
		for i := 0; i < len(buf); i++ {
			fmt.Printf("%2x ", buf[i])
		}
		fmt.Println("")

		// n, err = conn.Write(buf[:n])
		// if err != nil {
		// 	log.Printf("write failed: %v", err)
		// 	return err
		// }
		// log.Printf("write: %d", n)
	}
}

func (s *Server) sctpListen() (*sctp.SCTPListener, error) {
	ips := []net.IPAddr{}

	addr := &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    S1AP_PORT_NUMBER,
	}

	// Create SCTP server.
	return sctp.ListenSCTP("sctp", addr)
	// if err != nil {
	// 	log.Fatalf("failed to listen: %v", err)
	// }
	// fmt.Printf("Listen on %s\n", ln.Addr())
}

// Start() function initiate MME services.
func (s *Server) Start() error {
	mdata, err := asn1.Marshal(13)
	if err != nil {
		fmt.Println("asn1 error", err.Error())
	}
	fmt.Println("mdata", mdata)

	if s.done != nil {
		return fmt.Errorf("Server already started")
	}
	s.wg.Add(1)
	s.done = make(chan interface{})

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

				go s.serveClient(wconn)
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
