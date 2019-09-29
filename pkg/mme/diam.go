package mme

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/fiorix/go-diameter/diam/sm"
)

// DiamClient is S6A diameter protocol client.
type DiamClient struct {
	cli  *sm.Client
	conn diam.Conn
	wg   sync.WaitGroup
	opt  *DiamOpt
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
	} else {
		return opt.hssConnMethod
	}
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
					diam.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(opt.appID)),
					diam.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(opt.vendorID)),
				},
			}),
		},
	}

	return &DiamClient{
		cli: cli,
		opt: opt,
	}
}

// Start initiate diameter client start.
func (d *DiamClient) Start() {
	fmt.Println("Start")
	// m := diam.NewRequest(diam.AuthenticationInformation, diam.TGPP_S6A_APP_ID, dict.Default)
	// fmt.Println(m)

	go func() {
		for {
			fmt.Println("Trying to connect diameter server")
			// Right now DialNetworkTimeout will block until specified timeout.
			// There is no way to cancel it with something like Context mechanism.
			// So Stop() function may take up to timeout value.
			conn, err := d.cli.DialNetwork("tcp4", "172.16.0.52:3868")
			if err == nil {
				d.conn = conn
				break
			}
			fmt.Println("Failed", err)
			time.Sleep(time.Second * 3)
		}
		fmt.Println("Diam connection success!")
	}()
}

// Stop stops diameter client.
func (d *DiamClient) Stop() {
	// When connection is already established, close the connection.
	if d.conn != nil {
		d.conn.Close()
	}
	// d.wg.Wait()
}
