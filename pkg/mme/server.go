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

// ServerConfig keep MME server configuration.
type ServerConfig struct {
	retryTime time.Duration
}

// Server message.
type message struct {
	conn   net.Conn
	header []byte
	p      unsafe.Pointer
	typ    int
}

// Server is MME top level structure.
type Server struct {
	conf           ServerConfig
	ln             *sctp.SCTPListener
	wg             sync.WaitGroup
	ch             chan *message
	done           chan interface{}
	enb_ie_s1ap_id int32
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

func SCTPDumpBuf(buf []byte) {
	log.Printf("Packet length %d\n", len(buf))
	for i := 0; i < len(buf); i++ {
		fmt.Printf("%02x ", buf[i])
		if (i+1)%16 == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}

func SCTPBuffer() []byte {
	bufsize := 2048
	buf := make([]byte, bufsize+128) // Add overhead of SCTPSndRcvInfoWrappedConn
	return buf
}

func (s *Server) serveClient(conn net.Conn, infoSize int) error {
	for {
		buf := SCTPBuffer()

		n, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if n < infoSize {
			return fmt.Errorf("n (%d) < SCTPinfoSize (%d)", n, infoSize)
		}
		log.Printf("Read length: %d", n)

		header := buf[:infoSize]
		payload := buf[infoSize:n]
		SCTPDumpBuf(payload)

		p, typ, err := s1ap.Decode(payload)
		if err != nil {
			return fmt.Errorf("S1AP decode error")
		}
		s.ch <- &message{conn, header, p, typ}
	}
}

func (s *Server) sctpListen() (*sctp.SCTPListener, error) {
	ipaddr := net.IPAddr{
		IP: net.ParseIP("172.16.0.53"),
	}
	ips := []net.IPAddr{ipaddr}
	addr := &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    S1AP_PORT_NUMBER,
	}

	return sctp.ListenSCTP("sctp", addr)
}

func (s *Server) send(conn net.Conn, buf []byte) {
	fmt.Println("conn", conn)
	n, err := conn.Write(buf[:])
	if err != nil {
		log.Printf("write failed: %v", err)
	} else {
		log.Printf("write success %v bytes written!", n)
	}
}

// startHandler start S1AP packet handler.
func (s *Server) startHandler() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case msg := <-s.ch:
				s1ap.XerPrint(msg.p)
				switch msg.typ {
				case s1ap.S1_SETUP_REQUEST:
					log.Println("S1 SETUP REQUEST")
					payload, err := s1ap.S1SetupResponse()
					if err != nil {
						log.Println("S1SetupResponse error")
						continue
					}
					SCTPDumpBuf(payload)
					buf := append(msg.header, payload...)
					s.send(msg.conn, buf)
				case s1ap.INITIAL_UE_MESSAGE:
					log.Println("INITIAL UE MESSAGE")
					enb_ie_s1ap_id, err := s1ap.InitialUEMessageHandle(msg.p)
					if err != nil {
						log.Println("Initial UE Message error")
						continue
					}
					s.enb_ie_s1ap_id = enb_ie_s1ap_id
					mmebuf := []byte{
						0x07, 0x52, 0x00, 0x37, 0x74, 0x76, 0x61, 0x5c,
						0xb6, 0xd3, 0x7a, 0x91, 0x7d, 0x05, 0x72, 0x74,
						0x61, 0xb2, 0x41, 0x10, 0x7e, 0x0f, 0x9d, 0x7d,
						0x5a, 0xcb, 0x80, 0x00, 0x9f, 0xb3, 0xb3, 0x19,
						0x2a, 0x4c, 0x72, 0x12,
					}
					payload, err := s1ap.DownlinkNASTransport(enb_ie_s1ap_id, mmebuf)
					if err != nil {
						log.Println("DownlinkNASTransport error")
						continue
					}
					SCTPDumpBuf(payload)
					buf := append(msg.header, payload...)
					s.send(msg.conn, buf)
				case s1ap.UPLINK_NAS_TRANSPORT:
					_, eps_mmm_type, err := s1ap.UplinkNASTransportHandle(msg.p)
					if err != nil {
						continue
					}
					switch eps_mmm_type {
					case s1ap.NAS_EPS_AUTH_RESPONSE:
						mmebuf := []byte{
							0x37, 0x9f, 0x76, 0xaf, 0xd9, 0x00, 0x07, 0x5d,
							0x02, 0x00, 0x02, 0x80, 0x20,
						}
						payload, err := s1ap.DownlinkNASTransport(s.enb_ie_s1ap_id, mmebuf)
						if err != nil {
							log.Println("DownlinkNASTransport error")
							continue
						}
						SCTPDumpBuf(payload)
						buf := append(msg.header, payload...)
						s.send(msg.conn, buf)
					case s1ap.NAS_EPS_SECURITY_MODE_COMPLETE:
						payload, err := s1ap.InitialContextSetupRequest(s.enb_ie_s1ap_id)
						if err != nil {
							log.Println("InitialContextSetupRequest error")
							continue
						}
						SCTPDumpBuf(payload)
						buf := append(msg.header, payload...)
						s.send(msg.conn, buf)
					default:
						fmt.Println("Skip unknown MMM type")
					}
				default:
				}
				s1ap.Free(msg.p)
			case <-s.done:
				return
			}
		}
	}()
}

// startServer start SCTP server.
func (s *Server) startServer() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		infoSize := SCTPInfoSize()
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
}

// Start function initiate MME services.
func (s *Server) Start() error {
	diamOpt := &DiamOpt{
		originHost:       "mme.coreswitch.io",
		originRealm:      "coreswitch.io",
		vendorID:         10415,
		hostAddress:      "172.16.0.53",
		productName:      "coreswitch",
		firmwareRevision: 1,
		hssConnMethod:    "tcp4",
		hssAddress:       "172.16.0.52",
	}
	diam := NewDiamClient(diamOpt)
	diam.Start()
	// diam.Stop()

	// err = sendAIR(conn, cfg)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	///////////////////
	// SCTP S1AP Server.
	// if s.done != nil {
	// 	return fmt.Errorf("Server already started")
	// }
	// s.ch = make(chan *message, 1024)
	// s.done = make(chan interface{})

	// s.startHandler()
	// s.startServer()
	///////////////////

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
