package mme

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	log "github.com/coreswitch/log"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/fiorix/go-diameter/diam/sm"
	"github.com/fiorix/go-diameter/diam/sm/smpeer"
)

// DiamClient is S6A diameter protocol client.
type DiamClient struct {
	cli  *sm.Client
	opt  *DiamOpt
	cfg  *sm.Settings
	conn diam.Conn
	wg   sync.WaitGroup
}

// DiamOpt is DiamClient options.
type DiamOpt struct {
	// timeout
	originHost       string
	originRealm      string
	vendorID         uint32
	appID            uint32
	hostAddress      string
	productName      string
	firmwareRevision int
	watchdogInterval int
	hssConnMethod    string
	hssAddress       string
	hssPort          string
}

func (opt *DiamOpt) connMethod() string {
	if opt.hssConnMethod == "" {
		return "tcp4"
	}
	return opt.hssConnMethod
}

// AppID ...
func (opt *DiamOpt) AppID() uint32 {
	if opt.appID == 0 {
		return diam.TGPP_S6A_APP_ID
	}
	return opt.appID
}

// HssPort return HSSPort value.
func (opt *DiamOpt) HssPort() string {
	if opt.hssPort == "" {
		return "3868"
	}
	return opt.hssPort
}

// HssAddress ...
func (opt *DiamOpt) HssAddress() string {
	if opt.hssAddress == "" {
		return "127.0.0.1"
	}
	return opt.hssAddress
}

// NewDiamClient create new Diameter session for HSS.
func NewDiamClient(opt *DiamOpt) *DiamClient {
	cfg := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(opt.originHost),
		OriginRealm:      datatype.DiameterIdentity(opt.originRealm),
		VendorID:         datatype.Unsigned32(opt.vendorID),
		ProductName:      datatype.UTF8String(opt.productName),
		OriginStateID:    datatype.Unsigned32(time.Now().Unix()),
		FirmwareRevision: datatype.Unsigned32(opt.firmwareRevision),
		HostIPAddresses: []datatype.Address{
			datatype.Address(net.ParseIP(opt.hostAddress)),
		},
	}

	mux := sm.New(cfg)

	cli := &sm.Client{
		Dict:               dict.Default,
		Handler:            mux,
		MaxRetransmits:     0,
		RetransmitInterval: time.Second,
		EnableWatchdog:     true,
		WatchdogInterval:   time.Duration(opt.watchdogInterval) * time.Second,
		SupportedVendorID: []*diam.AVP{
			diam.NewAVP(avp.SupportedVendorID, avp.Mbit, 0, datatype.Unsigned32(opt.vendorID)),
		},
		VendorSpecificApplicationID: []*diam.AVP{
			diam.NewAVP(avp.VendorSpecificApplicationID, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(opt.AppID())),
					diam.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(opt.vendorID)),
				},
			}),
		},
	}

	mux.HandleIdx(diam.ALL_CMD_INDEX, handleAll())

	return &DiamClient{
		cli: cli,
		opt: opt,
		cfg: cfg,
	}
}

func handleAll() diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		log.Infof("Received Meesage From %s\n%s\n", c.RemoteAddr(), m)
	}
}

// sendAIR ...
func sendAIR(c diam.Conn, cfg *sm.Settings) (int64, error) {
	meta, ok := smpeer.FromContext(c.Context())
	if !ok {
		return 0, errors.New("peer metadata unavailable")
	}
	m := diam.NewRequest(diam.AuthenticationInformation, diam.TGPP_S6A_APP_ID, dict.Default)
	sid := "session;" + strconv.Itoa(int(rand.Uint32()))
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sid))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, meta.OriginRealm)
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, meta.OriginHost)

	ueIMSI := "001010000000001"
	m.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String(ueIMSI))

	m.NewAVP(avp.AuthSessionState, avp.Mbit, 0, datatype.Enumerated(0))

	plmnID := "\x00\xF1\x10"
	m.NewAVP(avp.VisitedPLMNID, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.OctetString(plmnID))

	m.NewAVP(avp.RequestedEUTRANAuthenticationInfo, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(
				avp.NumberOfRequestedVectors, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.Unsigned32(3)),
			diam.NewAVP(
				avp.ImmediateResponsePreferred, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.Unsigned32(0)),
		},
	})

	n, err := m.WriteTo(c)
	log.Infof("DIAM: %d written", n)

	return n, err
}

// sendCER - CapabilitiesExchange Reqeust for SCTP client.
func sendCER(c diam.Conn, cfg *sm.Settings) (int64, error) {
	m := diam.NewRequest(diam.CapabilitiesExchange, diam.TGPP_S6A_APP_ID, dict.Default)

	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)
	m.NewAVP(avp.OriginStateID, avp.Mbit, 0, cfg.OriginStateID)
	for _, addr := range cfg.HostIPAddresses {
		m.NewAVP(avp.HostIPAddress, avp.Mbit, 0, addr)
	}

	fmt.Println(m)

	n, err := m.WriteTo(c)
	log.Infof("DIAM: %d written", n)

	return n, err
}

// Start initiate diameter client start.
func (d *DiamClient) Start() {
	log.Info("Start")

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			log.Info("Trying to connect diameter server")

			conn, err := d.cli.DialNetwork(d.opt.connMethod(), d.opt.HssAddress()+":"+d.opt.HssPort())
			if err == nil {
				d.conn = conn
				break
			}
			log.Warnf("Failed %s", err)
			time.Sleep(time.Second * 3)
		}
		log.Info("Diam connection success!")
		sendAIR(d.conn, d.cfg)
	}()
}

// Stop stops diameter client.
func (d *DiamClient) Stop() {
	// When connection is already established, close the connection.
	if d.conn != nil {
		d.conn.Close()
	}
	d.wg.Wait()
}
