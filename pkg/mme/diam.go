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
	cli   *sm.Client
	opt   *DiamOpt
	cfg   *sm.Settings
	param *DiamParam
	done  chan struct{}
	conn  diam.Conn
	wg    sync.WaitGroup
}

// DiamParam store parameter for HSS session.
type DiamParam struct {
	ueIMSI string
	plmnID string
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

	done := make(chan struct{}, 1000)
	mux.HandleIdx(
		diam.CommandIndex{AppID: diam.TGPP_S6A_APP_ID, Code: diam.AuthenticationInformation, Request: false},
		handleAuthenticationInformationAnswer(done))
	mux.HandleIdx(
		diam.CommandIndex{AppID: diam.TGPP_S6A_APP_ID, Code: diam.UpdateLocation, Request: false},
		handleUpdateLocationAnswer(done))
	mux.HandleIdx(diam.ALL_CMD_INDEX, handleAll())

	return &DiamClient{
		cli:  cli,
		opt:  opt,
		cfg:  cfg,
		done: done,
	}
}

type ExperimentalResult struct {
	ExperimentalResultCode datatype.Unsigned32 `avp:"Experimental-Result-Code"`
}

type AuthenticationInfo struct {
	EUtranVector EUtranVector `avp:"E-UTRAN-Vector"`
}

type EUtranVector struct {
	RAND  datatype.OctetString `avp:"RAND"`
	XRES  datatype.OctetString `avp:"XRES"`
	AUTN  datatype.OctetString `avp:"AUTN"`
	KASME datatype.OctetString `avp:"KASME"`
}

type AIA struct {
	SessionID          datatype.UTF8String       `avp:"Session-Id"`
	ResultCode         datatype.Unsigned32       `avp:"Result-Code"`
	OriginHost         datatype.DiameterIdentity `avp:"Origin-Host"`
	OriginRealm        datatype.DiameterIdentity `avp:"Origin-Realm"`
	AuthSessionState   datatype.UTF8String       `avp:"Auth-Session-State"`
	ExperimentalResult ExperimentalResult        `avp:"Experimental-Result"`
	AIs                []AuthenticationInfo      `avp:"Authentication-Info"`
}

type AMBR struct {
	MaxRequestedBandwidthUL uint32 `avp:"Max-Requested-Bandwidth-UL"`
	MaxRequestedBandwidthDL uint32 `avp:"Max-Requested-Bandwidth-DL"`
}

type AllocationRetentionPriority struct {
	PriorityLevel           uint32 `avp:"Priority-Level"`
	PreemptionCapability    int32  `avp:"Pre-emption-Capability"`
	PreemptionVulnerability int32  `avp:"Pre-emption-Vulnerability"`
}

type EPSSubscribedQoSProfile struct {
	QoSClassIdentifier          int32                       `avp:"QoS-Class-Identifier"`
	AllocationRetentionPriority AllocationRetentionPriority `avp:"Allocation-Retention-Priority"`
}

type APNConfiguration struct {
	ContextIdentifier       uint32                  `avp:"Context-Identifier"`
	PDNType                 int32                   `avp:"PDN-Type"`
	ServiceSelection        string                  `avp:"Service-Selection"`
	EPSSubscribedQoSProfile EPSSubscribedQoSProfile `avp:"EPS-Subscribed-QoS-Profile"`
	AMBR                    AMBR                    `avp:"AMBR"`
}

type APNConfigurationProfile struct {
	ContextIdentifier                     uint32           `avp:"Context-Identifier"`
	AllAPNConfigurationsIncludedIndicator int32            `avp:"All-APN-Configurations-Included-Indicator"`
	APNConfiguration                      APNConfiguration `avp:"APN-Configuration"`
}

type SubscriptionData struct {
	MSISDN                        datatype.OctetString    `avp:"MSISDN"`
	AccessRestrictionData         uint32                  `avp:"Access-Restriction-Data"`
	SubscriberStatus              int32                   `avp:"Subscriber-Status"`
	NetworkAccessMode             int32                   `avp:"Network-Access-Mode"`
	AMBR                          AMBR                    `avp:"AMBR"`
	APNConfigurationProfile       APNConfigurationProfile `avp:"APN-Configuration-Profile"`
	SubscribedPeriodicRauTauTimer uint32                  `avp:"Subscribed-Periodic-RAU-TAU-Timer"`
}

type ULA struct {
	SessionID          string                    `avp:"Session-Id"`
	ULAFlags           uint32                    `avp:"ULA-Flags"`
	SubscriptionData   SubscriptionData          `avp:"Subscription-Data"`
	AuthSessionState   int32                     `avp:"Auth-Session-State"`
	ResultCode         uint32                    `avp:"Result-Code"`
	OriginHost         datatype.DiameterIdentity `avp:"Origin-Host"`
	OriginRealm        datatype.DiameterIdentity `avp:"Origin-Realm"`
	ExperimentalResult ExperimentalResult        `avp:"Experimental-Result"`
}

func handleAuthenticationInformationAnswer(done chan struct{}) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		log.Infof("Received Authentication-Information Answer from %s\n%s\n", c.RemoteAddr(), m)
		var aia AIA
		err := m.Unmarshal(&aia)
		if err != nil {
			log.Infof("AIA Unmarshal failed: %s", err)
		} else {
			log.Infof("Unmarshaled Authentication-Information Answer:\n%#+v\n", aia)
		}
		ok := struct{}{}
		done <- ok
	}
}

func handleUpdateLocationAnswer(done chan struct{}) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		log.Infof("Received Update-Location Answer from %s\n%s\n", c.RemoteAddr(), m)
		var ula ULA
		err := m.Unmarshal(&ula)
		if err != nil {
			log.Infof("ULA Unmarshal failed: %s", err)
		} else {
			log.Infof("Unmarshaled UL Answer:\n%#+v\n", ula)
		}
		ok := struct{}{}
		done <- ok
	}
}

func handleAll() diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		log.Infof("Received Meesage From %s\n%s\n", c.RemoteAddr(), m)
	}
}

// sendAIR ...
func sendAIR(c diam.Conn, cfg *sm.Settings, param *DiamParam) (int64, error) {
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
	m.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String(param.ueIMSI))
	m.NewAVP(avp.AuthSessionState, avp.Mbit, 0, datatype.Enumerated(0))
	m.NewAVP(avp.VisitedPLMNID, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.OctetString(param.plmnID))

	m.NewAVP(avp.RequestedEUTRANAuthenticationInfo, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(
				avp.NumberOfRequestedVectors, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.Unsigned32(3)),
			diam.NewAVP(
				avp.ImmediateResponsePreferred, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.Unsigned32(0)),
		},
	})

	return m.WriteTo(c)
}

// sendCER - CapabilitiesExchange Reqeust for SCTP client.
func sendCER(c diam.Conn, cfg *sm.Settings, param *DiamParam) (int64, error) {
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

const ULR_FLAGS = 1<<1 | 1<<5

// sendULR send Update-Location Request.
func sendULR(c diam.Conn, cfg *sm.Settings, param *DiamParam) (int64, error) {
	meta, ok := smpeer.FromContext(c.Context())
	if !ok {
		return 0, errors.New("peer metadata unavailable")
	}
	sid := "session;" + strconv.Itoa(int(rand.Uint32()))
	m := diam.NewRequest(diam.UpdateLocation, diam.TGPP_S6A_APP_ID, dict.Default)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sid))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, meta.OriginRealm)
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, meta.OriginHost)
	m.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String(param.ueIMSI))
	m.NewAVP(avp.AuthSessionState, avp.Mbit, 0, datatype.Enumerated(0))
	m.NewAVP(avp.RATType, avp.Mbit, uint32(cfg.VendorID), datatype.Enumerated(1004))
	m.NewAVP(avp.ULRFlags, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.Unsigned32(ULR_FLAGS))
	m.NewAVP(avp.VisitedPLMNID, avp.Vbit|avp.Mbit, uint32(cfg.VendorID), datatype.OctetString(param.plmnID))
	log.Infof("\nSending ULR to %s\n%s\n", c.RemoteAddr(), m)
	return m.WriteTo(c)
}

// Start initiate diameter client start.
func (d *DiamClient) Start() {
	log.Info("Start")

	d.param = &DiamParam{
		ueIMSI: "001010000000001",
		plmnID: "\x00\xF1\x10",
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
	retry:
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

		retryFunc := func() {
			log.Error("Authentication Information timeout")
			d.conn.Close()
			d.conn = nil
		}

		_, err := sendAIR(d.conn, d.cfg, d.param)
		if err != nil {
			retryFunc()
			goto retry
		}
		select {
		case <-d.done:
			log.Info("Authentication Information success")
		case <-time.After(10 * time.Second):
			log.Error("Authentication Information timeout")
			retryFunc()
			goto retry
		}
		_, err = sendULR(d.conn, d.cfg, d.param)
		if err != nil {
			retryFunc()
			goto retry
		}
		select {
		case <-d.done:
			log.Info("Update Location success")
		case <-time.After(10 * time.Second):
			log.Error("Update Location timeout")
			retryFunc()
			goto retry
		}
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
